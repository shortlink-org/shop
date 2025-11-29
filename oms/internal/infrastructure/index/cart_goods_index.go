package index

import (
	"sync"

	"github.com/google/uuid"
)

// CartGoodsIndex maintains an index of goods in carts
// Maps good_id -> set of customer_ids that have this good in their cart
type CartGoodsIndex struct {
	mu    sync.RWMutex
	index map[uuid.UUID]map[uuid.UUID]bool // good_id -> customer_id -> true
}

// NewCartGoodsIndex creates a new cart goods index
func NewCartGoodsIndex() *CartGoodsIndex {
	return &CartGoodsIndex{
		index: make(map[uuid.UUID]map[uuid.UUID]bool),
	}
}

// AddGoodToCart adds a good to a customer's cart in the index
func (i *CartGoodsIndex) AddGoodToCart(goodId, customerId uuid.UUID) {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.index[goodId] == nil {
		i.index[goodId] = make(map[uuid.UUID]bool)
	}
	i.index[goodId][customerId] = true
}

// RemoveGoodFromCart removes a good from a customer's cart in the index
func (i *CartGoodsIndex) RemoveGoodFromCart(goodId, customerId uuid.UUID) {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.index[goodId] != nil {
		delete(i.index[goodId], customerId)
		if len(i.index[goodId]) == 0 {
			delete(i.index, goodId)
		}
	}
}

// GetCustomersWithGood returns all customer IDs that have the specified good in their cart
func (i *CartGoodsIndex) GetCustomersWithGood(goodId uuid.UUID) []uuid.UUID {
	i.mu.RLock()
	defer i.mu.RUnlock()

	customers := make([]uuid.UUID, 0, len(i.index[goodId]))
	for customerId := range i.index[goodId] {
		customers = append(customers, customerId)
	}
	return customers
}

// ClearCart clears all goods for a customer from the index
func (i *CartGoodsIndex) ClearCart(customerId uuid.UUID) {
	i.mu.Lock()
	defer i.mu.Unlock()

	for goodId, customers := range i.index {
		delete(customers, customerId)
		if len(customers) == 0 {
			delete(i.index, goodId)
		}
	}
}

