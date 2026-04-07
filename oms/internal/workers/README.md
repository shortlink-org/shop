# Temporal Workers

This directory contains Temporal workflow workers for OMS.

## Overview

OMS uses [Temporal](https://temporal.io/) for durable workflow orchestration. The **order** workflow exposes signals and a Query for status. The **cart** workflow orchestrates commands via signals and activities that persist through use cases; it does **not** expose a Temporal Query (cart reads go through the application layer and repository — see [cart use cases README](../usecases/cart/README.md)).

## Structure

```text
workers/
├── cart/
│   ├── workflow/
│   │   ├── workflow.go       # Cart workflow definition
│   │   ├── dto/              # Domain to workflow model mapping
│   │   └── model/            # Workflow-specific proto models
│   ├── cart_worker/
│   │   └── worker.go         # Temporal worker registration
│   └── di/                   # Dependency injection
└── order/
    ├── workflow/
    │   ├── workflow.go       # Order workflow definition
    │   ├── dto/
    │   └── model/
    ├── order_worker/
    │   └── worker.go
    └── di/
```

## Cart Workflow

Long-running workflow that reacts to signals by running **activities** which call the same cart command use cases used by gRPC, persisting state in **PostgreSQL**. The workflow does not keep a separate in-memory cart snapshot and **does not register a Query handler** — Temporal Queries cannot run activities or hit the database, so reads are not modeled here.

### Cart Signals

| Signal         | Description                |
|----------------|----------------------------|
| `EVENT_ADD`    | Add items to cart          |
| `EVENT_REMOVE` | Remove items from cart     |
| `EVENT_RESET`  | Reset cart to empty state  |

### Cart Queries

None. Use the gRPC `Get` path (repository-backed) or extend the workflow with in-memory state if you intentionally want `SetQueryHandler` (not the current design).

### Cart Sequence Diagram

```plantuml
@startuml Cart_Workflow
participant Client
participant "Temporal" as T
participant "Cart Workflow" as CW
participant "Activity" as A
database "PostgreSQL" as DB

== Start Workflow ==
Client -> T: StartWorkflow(customerId)
T -> CW: Initialize

== Add Items ==
Client -> T: Signal(EVENT_ADD, payload)
T -> CW: Receive signal
CW -> A: ExecuteActivity(AddItem)
A -> DB: Use case persists cart
A --> CW: success

== Remove / Reset ==
Client -> T: Signal(EVENT_REMOVE / EVENT_RESET, ...)
T -> CW: Receive signal
CW -> A: ExecuteActivity(RemoveItem / ResetCart)
A -> DB: Use case updates cart
A --> CW: success

== Read cart (not via this workflow) ==
note right of Client
  gRPC Get -> query use case -> CartRepository
  (no QueryWorkflow for cart)
end note

@enduml
```

## Order Workflow

Long-running workflow that manages order lifecycle.

### Order Signals

| Signal     | Description               |
|------------|---------------------------|
| `CANCEL`   | Cancel the order          |
| `COMPLETE` | Mark order as completed   |

### Order Queries

| Query | Description              |
|-------|--------------------------|
| `GET` | Get current order state  |

### Order Sequence Diagram

```plantuml
@startuml Order_Workflow
participant Client
participant "Temporal" as T
participant "Order Workflow" as OW
database "Workflow State" as WS

== Create Order ==
Client -> T: StartWorkflow(orderId, customerId, items)
T -> OW: Initialize
OW -> WS: Create order with items

== Get Order ==
Client -> T: Query(GET)
T -> OW: Execute query
OW --> T: Return order state
T --> Client: OrderState

== Complete Order ==
Client -> T: Signal(COMPLETE)
T -> OW: Receive signal
OW -> WS: Mark as completed

== Cancel Order (alternative) ==
Client -> T: Signal(CANCEL)
T -> OW: Receive signal
OW -> WS: Mark as cancelled

@enduml
```

## Task Queues

| Queue            | Temporal Name        | Purpose                    |
|------------------|----------------------|----------------------------|
| `CartTaskQueue`  | `CART_TASK_QUEUE`    | Cart workflow execution    |
| `OrderTaskQueue` | `ORDER_TASK_QUEUE`   | Order workflow execution   |

## Why Temporal?

See [ADR-0003](../../../docs/ADR/decisions/0003-temporal.md) for the decision rationale.

## Running Workers

Workers are started as part of the OMS service. Each worker polls its respective task queue for workflow tasks.

```go
// Cart worker
w := worker.New(c, temporal.GetQueueName(v1.CartTaskQueue), worker.Options{})
w.RegisterWorkflow(cart_workflow.Workflow)

// Order worker  
w := worker.New(c, temporal.GetQueueName(v1.OrderTaskQueue), worker.Options{})
w.RegisterWorkflow(order_workflow.Workflow)
```
