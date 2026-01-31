package dto

import (
	"github.com/shopspring/decimal"

	cart "github.com/shortlink-org/shop/oms/internal/domain/cart/v1"
	itemv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/item/v1"
	itemsv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/items/v1"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/repository/postgres/cart/schema/crud"
)

// ToDomain converts database models to domain aggregate.
func ToDomain(row crud.OmsCart, items []crud.GetCartItemsRow) *cart.State {
	domainItems := make(itemsv1.Items, 0, len(items))

	for _, i := range items {
		// Create item with pricing (tax is stored as 0 since we don't have it in DB schema)
		item, err := itemv1.NewItemWithPricing(i.GoodID, i.Quantity, i.Price, i.Discount, decimal.Zero)
		if err != nil {
			// Skip invalid items (should not happen with valid DB data)
			continue
		}

		domainItems = append(domainItems, item)
	}

	return cart.Reconstitute(row.CustomerID, domainItems, int(row.Version))
}
