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
service := services.NewStockCartService(cartRepository)
results, err := service.HandleStockDepletion(ctx, goodId, affectedCustomerIds)
```

### CartValidationService

Validates cart operations before execution.

**Use case:** Before adding items to cart, validate stock availability and quantity limits.

```go
service := cart_validation.New(stockChecker)
result := service.ValidateAddItems(ctx, items)
if !result.Valid {
    // Handle validation errors
}
```

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
