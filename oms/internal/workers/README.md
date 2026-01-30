# Temporal Workers

This directory contains Temporal workflow workers for OMS.

## Overview

OMS uses [Temporal](https://temporal.io/) for durable workflow orchestration. Each aggregate (Cart, Order) has its own long-running workflow that manages state through signals and queries.

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

Long-running workflow that manages shopping cart state.

### Cart Signals

| Signal         | Description                |
|----------------|----------------------------|
| `EVENT_ADD`    | Add items to cart          |
| `EVENT_REMOVE` | Remove items from cart     |
| `EVENT_RESET`  | Reset cart to empty state  |

### Cart Queries

| Query       | Description            |
|-------------|------------------------|
| `EVENT_GET` | Get current cart state |

### Cart Sequence Diagram

```plantuml
@startuml Cart_Workflow
participant Client
participant "Temporal" as T
participant "Cart Workflow" as CW
database "Workflow State" as WS

== Start Workflow ==
Client -> T: StartWorkflow(customerId)
T -> CW: Initialize
CW -> WS: Create empty cart

== Add Items ==
Client -> T: Signal(EVENT_ADD, items)
T -> CW: Receive signal
CW -> WS: Update cart state

== Get Cart ==
Client -> T: Query(EVENT_GET)
T -> CW: Execute query
CW --> T: Return cart state
T --> Client: CartState

== Remove Items ==
Client -> T: Signal(EVENT_REMOVE, items)
T -> CW: Receive signal
CW -> WS: Update cart state

== Reset Cart ==
Client -> T: Signal(EVENT_RESET)
T -> CW: Receive signal
CW -> WS: Reset to empty

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
