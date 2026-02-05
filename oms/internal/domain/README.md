# OMS Domain Layer

## Structure

```
domain/
├── cart/v1/                    # Cart Aggregate
│   ├── state.go                # Aggregate root
│   ├── add_item.go             # Aggregate methods
│   ├── remove_item.go
│   ├── vo/                     # Value Objects
│   │   ├── address/
│   │   ├── location/
│   │   └── weight/
│   ├── item/v1/                # Entity
│   ├── items/v1/               # Collection
│   └── events/v1/              # Domain Events
│
├── order/v1/                   # Order Aggregate
│   ├── order_state.go          # Aggregate root
│   ├── workflow_operations.go  # Aggregate methods
│   ├── vo/                     # Value Objects
│   └── events/v1/              # Domain Events (proto)
│
├── services/                   # Domain Services
│   ├── stock_cart_service.go   # Cart + Stock operations
│   └── cart_validation/v1/     # Cart validation rules
│
├── pricing/                    # Pricing domain
│   └── price_policy.go
│
├── queue/v1/                   # Queue domain
│   └── queue.go
│
└── stock/v1/                   # Stock domain
    └── stock_event.proto
```

## Aggregates

### Cart (`cart/v1/`)

Manages shopping cart state and operations.

- **State**: `state.go` - Cart aggregate root
- **Methods**: `AddItem()`, `RemoveItem()`, `Reset()`
- **Events**: `ItemAddedEvent`, `ItemRemovedEvent`, `ResetEvent`

### Order (`order/v1/`)

Manages order lifecycle.

- **State**: `order_state.go` - Order aggregate root
- **Methods**: Workflow operations for order processing
- **Events**: Order created, confirmed, shipped, etc.

## Domain Services

See [services/README.md](./services/README.md) for details.

- **StockCartService** - Handles stock depletion impact on carts
- **CartValidationService** - Validates cart operations

## Value Objects

Immutable objects defined by their attributes:

- `vo/address/` - Physical address
- `vo/location/` - GPS coordinates
- `vo/weight/` - Weight with units

## Domain Events

Events that represent state changes:

- Cart: `ItemAddedEvent`, `ItemRemovedEvent`, `ResetEvent`
- Order: Proto-defined events in `events/v1/`
