package dto

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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
		goodID := pgtypeUUIDToUUID(i.GoodID)
		price := pgtypeNumericToDecimal(i.Price)
		discount := pgtypeNumericToDecimal(i.Discount)

		// Create item with pricing (tax is stored as 0 since we don't have it in DB schema)
		item, err := itemv1.NewItemWithPricing(goodID, i.Quantity, price, discount, decimal.Zero)
		if err != nil {
			// Skip invalid items (should not happen with valid DB data)
			continue
		}

		domainItems = append(domainItems, item)
	}

	customerID := pgtypeUUIDToUUID(row.CustomerID)

	return cart.Reconstitute(customerID, domainItems, int(row.Version))
}

// pgtypeUUIDToUUID converts pgtype.UUID to uuid.UUID
func pgtypeUUIDToUUID(p pgtype.UUID) uuid.UUID {
	if !p.Valid {
		return uuid.Nil
	}
	return uuid.UUID(p.Bytes)
}

// pgtypeNumericToDecimal converts pgtype.Numeric to decimal.Decimal
func pgtypeNumericToDecimal(p pgtype.Numeric) decimal.Decimal {
	if !p.Valid {
		return decimal.Zero
	}

	// Convert pgtype.Numeric to float64
	f, err := p.Float64Value()
	if err == nil && f.Valid {
		return decimal.NewFromFloat(f.Float64)
	}

	// Fallback: try Int64
	i, err := p.Int64Value()
	if err == nil && i.Valid {
		return decimal.NewFromInt(i.Int64)
	}

	return decimal.Zero
}
