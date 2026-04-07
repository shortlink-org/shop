# Cart Use Cases

## Overview

The Cart domain manages shopping cart operations for customers. Each customer has exactly one cart. **Persistence and the public gRPC API go through application use cases and `CartRepository` (PostgreSQL)** — not through Temporal Queries.

A **Temporal cart workflow** exists in `internal/workers/cart/workflow` for optional orchestration (signals → activities → the same command handlers). It does **not** hold a copy of cart items in workflow memory and **does not register a Temporal Query handler**; reads therefore cannot be served by `QueryWorkflow`. The domain enum value `EVENT_GET` is reserved in protobuf but is **not** wired to a Temporal Query in this codebase.

The cart supports adding, removing, and reading items, as well as automatic cleanup when stock is depleted.

### Key Features

- **Repository-backed state**: Add/Remove/Reset/Get via use cases and `CartRepository`
- **Optional Temporal workflow**: Long-running session TTL and signal-driven activities (see [workers README](../../workers/README.md)); same handlers as gRPC, no `QueryWorkflow` for cart
- **Stock Integration**: Automatic item removal when stock is depleted (see [STOCK_CHANGES.md](STOCK_CHANGES.md))

## Architecture

```mermaid
graph TB
    subgraph clients [Clients]
        UI[Shop UI]
        API[External API]
    end

    subgraph grpc [gRPC Layer]
        CartRPC[CartService gRPC]
    end

    subgraph usecases [Use Cases]
        CartUC[Cart UseCase]
    end

    subgraph persistence [Persistence]
        CartRepo[CartRepository]
    end

    subgraph temporal [Temporal optional]
        CartWF[Cart Workflow]
        CartWorker[Cart Worker]
    end

    subgraph domain [Domain]
        Cart[Cart Aggregate]
        CartItem[CartItem Entity]
        Events[Domain Events]
    end

    UI --> CartRPC
    API --> CartRPC
    CartRPC --> CartUC
    CartUC --> CartRepo
    CartRepo --> Cart
    Cart --> CartItem
    Cart --> Events
    CartWorker --> CartWF
```

## State Machine

The cart operates as a continuous state machine where all operations are valid from any state:

```mermaid
stateDiagram-v2
    [*] --> Active: Customer Created
    
    state Active {
        [*] --> Empty
        Empty --> HasItems: Add Items
        HasItems --> HasItems: Add Items
        HasItems --> HasItems: Remove Items
        HasItems --> Empty: Remove All / Reset
        Empty --> Empty: Reset
    }
    
    Active --> [*]: Customer Deleted
```

## API Reference

### gRPC Service

**Package**: `infrastructure.rpc.cart.v1`

```protobuf
service CartService {
  rpc Add(AddRequest) returns (google.protobuf.Empty);
  rpc Remove(RemoveRequest) returns (google.protobuf.Empty);
  rpc Get(GetRequest) returns (GetResponse);
  rpc Reset(ResetRequest) returns (google.protobuf.Empty);
}
```

### Add Items

Adds one or more items to the customer's cart. If an item already exists, quantities are summed.

**Request:**
```json
{
  "customer_id": "550e8400-e29b-41d4-a716-446655440000",
  "items": [
    { "good_id": "123e4567-e89b-12d3-a456-426614174000", "quantity": 2 },
    { "good_id": "987fcdeb-51a2-3bc4-d567-890123456789", "quantity": 1 }
  ]
}
```

**Response:** `Empty`

### Remove Items

Removes items from the cart. If quantity becomes zero or negative, the item is removed entirely.

**Request:**
```json
{
  "customer_id": "550e8400-e29b-41d4-a716-446655440000",
  "items": [
    { "good_id": "123e4567-e89b-12d3-a456-426614174000", "quantity": 1 }
  ]
}
```

**Response:** `Empty`

### Get Cart

Retrieves the current cart state for a customer.

**Request:**
```json
{
  "customer_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Response:**
```json
{
  "state": {
    "cart_id": "cart-550e8400",
    "customer_id": "550e8400-e29b-41d4-a716-446655440000",
    "items": [
      { "good_id": "123e4567-e89b-12d3-a456-426614174000", "quantity": 2 }
    ],
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T11:45:00Z"
  }
}
```

### Reset Cart

Clears all items from the cart.

**Request:**
```json
{
  "customer_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Response:** `Empty`

## Temporal Workflow

### Workflow Details

| Property | Value |
|----------|-------|
| Task Queue | `CART_TASK_QUEUE` |
| Workflow ID | `cart-{customer_id}` |
| Lifetime | Long-running (until customer deleted) |

### Signals

| Signal | Payload | Description |
|--------|---------|-------------|
| `EVENT_ADD` | `CartEvent` | Add items to cart |
| `EVENT_REMOVE` | `CartEvent` | Remove items from cart |
| `EVENT_RESET` | `string` (customer_id) | Clear cart |

### Temporal Queries

The cart workflow **does not** expose a Temporal Query (no `SetQueryHandler`). Reading the cart for the gRPC API uses `query/get` → `CartRepository`, not `client.QueryWorkflow`.

The protobuf enum `domain.cart.v1.Event.EVENT_GET` is not used as a Temporal Query name in the implementation.

### Sequence Diagrams (gRPC — default path)

```mermaid
sequenceDiagram
    participant Client
    participant gRPC as CartService
    participant UC as CartUseCase
    participant Repo as CartRepository

    Note over Client,Repo: Add Items
    Client->>gRPC: Add(AddRequest)
    gRPC->>UC: add_items.Handler.Handle
    UC->>Repo: Load / Save cart aggregate
    UC-->>gRPC: success
    gRPC-->>Client: Empty

    Note over Client,Repo: Get Cart
    Client->>gRPC: Get(GetRequest)
    gRPC->>UC: get.Handler.Handle
    UC->>Repo: Load cart aggregate
    Repo-->>UC: Cart state
    UC-->>gRPC: GetResponse
    gRPC-->>Client: GetResponse
```

## Error Handling

### Error Codes

| Code | Description | Recovery |
|------|-------------|----------|
| `INVALID_ARGUMENT` | Invalid customer_id or good_id format | Fix UUID format |
| `NOT_FOUND` | Resource missing (context-dependent) | Verify IDs / permissions |
| `INTERNAL` | Temporal communication error | Retry with backoff |

### Domain Validation Rules

Items are validated before being added to cart:

| Rule | Description |
|------|-------------|
| `GoodIdNotEmpty` | good_id must be a valid UUID |
| `QuantityPositive` | quantity must be > 0 |
| `PriceNonNegative` | price must be >= 0 |
| `DiscountNonNegative` | discount must be >= 0 |
| `DiscountNotExceedsPrice` | discount <= price |
| `TaxNonNegative` | tax must be >= 0 |

## Integration

### Stock Changes

When inventory reaches zero, items are automatically removed from all carts containing that product. See [STOCK_CHANGES.md](STOCK_CHANGES.md) for details.

### WebSocket Notifications

Customers receive real-time notifications when items are removed due to stock depletion:

```javascript
const ws = new WebSocket(`ws://oms:50051/ws?customer_id=${customerId}`);
ws.onmessage = (event) => {
  const { type, data } = JSON.parse(event.data);
  if (type === 'stock_depleted') {
    // Refresh cart UI
  }
};
```

### Creating Orders from Cart

The Order service can create an order from the current cart state:

```go
// Get cart state
cartState := cartUC.Get(ctx, customerId)

// Create order from cart items
order := orderUC.CreateFromCart(ctx, cartState)
```
