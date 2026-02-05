# Domain Services

Domain services contain business logic that:

- Operates on multiple aggregates (e.g., Cart and Stock)
- Doesn't belong to a single entity
- Has no infrastructure dependencies

## Services

### StockCartService

Handles stock-related operations for carts.

**Use case:** When stock for a good is depleted, remove the item from all affected carts.

```go
service := services.NewStockCartService()
// Use case: Load cart -> service.ProcessStockDepletion(cart, goodId) -> Save cart
result := service.ProcessStockDepletion(cartState, goodId)
```

### CartValidationService

Validates cart operations before execution. **No I/O**: the domain only interprets pre-fetched data. The use case must obtain stock data via a port (e.g. StockChecker) and pass it in.

**Use case:** Before adding items to cart, the use case fetches stock for each item, then calls the pure validation:

```go
service := cart_validation.New()
// Use case: fetch stockByGoodId via StockChecker port, then:
stockByGoodId := make(map[uuid.UUID]cart_validation.StockAvailabilityInput)
for _, item := range items {
    available, qty, err := stockChecker.CheckStockAvailability(ctx, item.GetGoodId(), item.GetQuantity())
    stockByGoodId[item.GetGoodId()] = cart_validation.StockAvailabilityInput{
        GoodID: item.GetGoodId(), Available: available, StockQuantity: qty, CheckError: err,
    }
}
result := service.ValidateAddItems(items, stockByGoodId)
if !result.Valid {
    // Handle validation errors
}
```

Or use the package-level pure function: `cart_validation.ValidateAddItemsWithStock(items, stockByGoodId)`.

## Domain Services vs Use Cases

| Aspect          | Domain Services             | Use Cases                        |
|-----------------|-----------------------------|---------------------------------|
| **Location**    | `domain/services/`          | `usecases/`                     |
| **Dependencies**| Only domain objects         | Repositories, external services |
| **I/O**         | None                        | Database, HTTP, messaging       |
| **Purpose**     | Pure business logic         | Orchestration                   |

## When to Use Domain Services

Use domain services when:

1. Logic involves multiple aggregates
2. Business rules don't belong to a single entity
3. You need pure, testable business logic

Do NOT use for:

1. Simple operations on a single aggregate (use aggregate methods)
2. Infrastructure operations (use infrastructure layer)
3. Workflow coordination (use use cases)
