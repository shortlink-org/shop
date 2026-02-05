package cart_goods_index

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/redis/rueidis"
)

const (
	// KeyPrefix is the prefix for all cart goods index keys
	keyPrefix = "oms:cart:good"
	// CustomerGoodsPrefix is the prefix for reverse index (customer -> goods)
	customerGoodsPrefix = "oms:cart:customer"
)

// Store implements ports.CartGoodsIndex using Redis.
type Store struct {
	client rueidis.Client
}

// New creates a new Redis CartGoodsIndex.
func New(client rueidis.Client) *Store {
	return &Store{client: client}
}

// goodCustomersKey returns the key for storing customers who have a specific good.
// Pattern: oms:cart:good:{good_id}:customers
func goodCustomersKey(goodID uuid.UUID) string {
	return fmt.Sprintf("%s:%s:customers", keyPrefix, goodID.String())
}

// customerGoodsKey returns the key for storing goods in a customer's cart.
// Pattern: oms:cart:customer:{customer_id}:goods
// This reverse index is needed for ClearCart operation.
func customerGoodsKey(customerID uuid.UUID) string {
	return fmt.Sprintf("%s:%s:goods", customerGoodsPrefix, customerID.String())
}

// AddGoodToCart adds a good to a customer's cart in the index.
// Uses two Redis SETs for bidirectional lookup:
// - good -> customers (for GetCustomersWithGood)
// - customer -> goods (for ClearCart)
func (s *Store) AddGoodToCart(ctx context.Context, goodID, customerID uuid.UUID) error {
	// Add customer to good's customer set
	cmd1 := s.client.B().Sadd().Key(goodCustomersKey(goodID)).Member(customerID.String()).Build()
	// Add good to customer's goods set (reverse index)
	cmd2 := s.client.B().Sadd().Key(customerGoodsKey(customerID)).Member(goodID.String()).Build()

	// Execute both commands
	for _, resp := range s.client.DoMulti(ctx, cmd1, cmd2) {
		err := resp.Error()
		if err != nil {
			return fmt.Errorf("failed to add good to cart index: %w", err)
		}
	}

	return nil
}

// RemoveGoodFromCart removes a good from a customer's cart in the index.
func (s *Store) RemoveGoodFromCart(ctx context.Context, goodID, customerID uuid.UUID) error {
	// Remove customer from good's customer set
	cmd1 := s.client.B().Srem().Key(goodCustomersKey(goodID)).Member(customerID.String()).Build()
	// Remove good from customer's goods set
	cmd2 := s.client.B().Srem().Key(customerGoodsKey(customerID)).Member(goodID.String()).Build()

	// Execute both commands
	for _, resp := range s.client.DoMulti(ctx, cmd1, cmd2) {
		err := resp.Error()
		if err != nil {
			return fmt.Errorf("failed to remove good from cart index: %w", err)
		}
	}

	return nil
}

// GetCustomersWithGood returns all customer IDs that have the specified good in their cart.
func (s *Store) GetCustomersWithGood(ctx context.Context, goodID uuid.UUID) ([]uuid.UUID, error) {
	members, err := s.client.Do(ctx,
		s.client.B().Smembers().Key(goodCustomersKey(goodID)).Build(),
	).AsStrSlice()
	if err != nil {
		// Empty set returns error, treat as empty result
		if rueidis.IsRedisNil(err) {
			return []uuid.UUID{}, nil
		}

		return nil, fmt.Errorf("failed to get customers with good: %w", err)
	}

	customers := make([]uuid.UUID, 0, len(members))
	for _, member := range members {
		id, err := uuid.Parse(member)
		if err != nil {
			// Skip invalid UUIDs (shouldn't happen, but be defensive)
			continue
		}

		customers = append(customers, id)
	}

	return customers, nil
}

// ClearCart removes all goods for a customer from the index.
// This uses the reverse index to find all goods and remove the customer from each.
func (s *Store) ClearCart(ctx context.Context, customerID uuid.UUID) error {
	// Get all goods in customer's cart
	goods, err := s.client.Do(ctx,
		s.client.B().Smembers().Key(customerGoodsKey(customerID)).Build(),
	).AsStrSlice()
	if err != nil {
		if rueidis.IsRedisNil(err) {
			return nil // No goods to clear
		}

		return fmt.Errorf("failed to get customer goods: %w", err)
	}

	if len(goods) == 0 {
		return nil
	}

	// Build commands to remove customer from each good's customer set
	cmds := make([]rueidis.Completed, 0, len(goods)+1)
	for _, goodIDStr := range goods {
		goodID, err := uuid.Parse(goodIDStr)
		if err != nil {
			continue
		}

		cmds = append(cmds, s.client.B().Srem().Key(goodCustomersKey(goodID)).Member(customerID.String()).Build())
	}

	// Delete the customer's goods set
	cmds = append(cmds, s.client.B().Del().Key(customerGoodsKey(customerID)).Build())

	// Execute all commands
	for _, resp := range s.client.DoMulti(ctx, cmds...) {
		err := resp.Error()
		if err != nil {
			return fmt.Errorf("failed to clear cart index: %w", err)
		}
	}

	return nil
}
