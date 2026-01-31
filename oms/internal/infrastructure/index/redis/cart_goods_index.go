package redis

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/redis/rueidis"
)

const (
	// keyPrefix is the prefix for all cart goods index keys
	keyPrefix = "oms:cart:good"
	// customerGoodsPrefix is the prefix for reverse index (customer -> goods)
	customerGoodsPrefix = "oms:cart:customer"
)

// CartGoodsIndex implements ports.CartGoodsIndex using Redis.
type CartGoodsIndex struct {
	client rueidis.Client
}

// New creates a new Redis CartGoodsIndex.
func New(client rueidis.Client) *CartGoodsIndex {
	return &CartGoodsIndex{client: client}
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
func (i *CartGoodsIndex) AddGoodToCart(ctx context.Context, goodID, customerID uuid.UUID) error {
	// Add customer to good's customer set
	cmd1 := i.client.B().Sadd().Key(goodCustomersKey(goodID)).Member(customerID.String()).Build()
	// Add good to customer's goods set (reverse index)
	cmd2 := i.client.B().Sadd().Key(customerGoodsKey(customerID)).Member(goodID.String()).Build()

	// Execute both commands
	for _, resp := range i.client.DoMulti(ctx, cmd1, cmd2) {
		if err := resp.Error(); err != nil {
			return fmt.Errorf("failed to add good to cart index: %w", err)
		}
	}

	return nil
}

// RemoveGoodFromCart removes a good from a customer's cart in the index.
func (i *CartGoodsIndex) RemoveGoodFromCart(ctx context.Context, goodID, customerID uuid.UUID) error {
	// Remove customer from good's customer set
	cmd1 := i.client.B().Srem().Key(goodCustomersKey(goodID)).Member(customerID.String()).Build()
	// Remove good from customer's goods set
	cmd2 := i.client.B().Srem().Key(customerGoodsKey(customerID)).Member(goodID.String()).Build()

	// Execute both commands
	for _, resp := range i.client.DoMulti(ctx, cmd1, cmd2) {
		if err := resp.Error(); err != nil {
			return fmt.Errorf("failed to remove good from cart index: %w", err)
		}
	}

	return nil
}

// GetCustomersWithGood returns all customer IDs that have the specified good in their cart.
func (i *CartGoodsIndex) GetCustomersWithGood(ctx context.Context, goodID uuid.UUID) ([]uuid.UUID, error) {
	members, err := i.client.Do(ctx,
		i.client.B().Smembers().Key(goodCustomersKey(goodID)).Build(),
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
func (i *CartGoodsIndex) ClearCart(ctx context.Context, customerID uuid.UUID) error {
	// Get all goods in customer's cart
	goods, err := i.client.Do(ctx,
		i.client.B().Smembers().Key(customerGoodsKey(customerID)).Build(),
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
		cmds = append(cmds, i.client.B().Srem().Key(goodCustomersKey(goodID)).Member(customerID.String()).Build())
	}

	// Delete the customer's goods set
	cmds = append(cmds, i.client.B().Del().Key(customerGoodsKey(customerID)).Build())

	// Execute all commands
	for _, resp := range i.client.DoMulti(ctx, cmds...) {
		if err := resp.Error(); err != nil {
			return fmt.Errorf("failed to clear cart index: %w", err)
		}
	}

	return nil
}
