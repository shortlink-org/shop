# 3. Use Temporal for Workflow Orchestration

Date: 2024-05-05

## Status

Accepted

## Context

OMS needs to manage long-running business processes:

- **Cart management** - cart state must persist across user sessions
- **Order processing** - orders go through multiple states (created → confirmed → shipped → delivered)
- **Saga pattern** - distributed transactions across services

Requirements:

- Durable state persistence
- Fault tolerance and automatic retries
- Visibility into workflow state
- Support for long-running processes (hours/days)

## Decision

Use [Temporal](https://temporal.io/) for workflow orchestration.

### Alternatives Considered

| Option                     | Pros                                       | Cons                                     |
|----------------------------|--------------------------------------------|------------------------------------------|
| **Temporal**               | Durable execution, built-in retries        | Additional infrastructure                |
| **Custom state machine**   | Simple, no dependencies                    | No durability, manual retry logic        |
| **Kafka + event sourcing** | Event-driven, scalable                     | Complex replay logic                     |
| **Database + cron**        | Simple polling                             | Not real-time, scaling issues            |

### Why Temporal

1. **Durable Execution** - workflow state survives service restarts
2. **Fault Tolerance** - automatic retries with configurable policies
3. **Visibility** - UI for monitoring workflows
4. **Developer Experience** - write workflows as regular code
5. **Scalability** - horizontal scaling of workers

## Implementation

### Workflows

Each aggregate has its own workflow:

```plantuml
@startuml Temporal_Architecture
!include https://raw.githubusercontent.com/plantuml-stdlib/C4-PlantUML/master/C4_Container.puml

LAYOUT_WITH_LEGEND()

title Temporal Workflow Architecture

System_Boundary(oms, "OMS Service") {
    Container(api, "gRPC API", "Go", "Handles requests")
    Container(cart_worker, "Cart Worker", "Go", "Executes cart workflows")
    Container(order_worker, "Order Worker", "Go", "Executes order workflows")
}

System_Ext(temporal, "Temporal Server", "Orchestration")
ContainerDb(temporal_db, "Temporal DB", "PostgreSQL", "Workflow state")

Rel(api, temporal, "Start/Signal/Query workflows", "gRPC")
Rel(temporal, cart_worker, "Dispatch tasks", "gRPC")
Rel(temporal, order_worker, "Dispatch tasks", "gRPC")
Rel(temporal, temporal_db, "Persist state", "SQL")

@enduml
```

### Communication Pattern

```text
Client → gRPC API → Temporal Client → Temporal Server → Worker → Workflow
                         ↑                                   ↓
                         └─────────── Query/Signal ──────────┘
```

### Task Queues

| Workflow | Task Queue           |
|----------|----------------------|
| Cart     | `CART_TASK_QUEUE`    |
| Order    | `ORDER_TASK_QUEUE`   |

## Consequences

### Positive

- Reliable order processing even during failures
- Easy to add new workflow steps
- Built-in observability via Temporal UI
- Versioning support for workflow updates

### Negative

- Additional infrastructure to maintain (Temporal Server)
- Learning curve for team
- Debugging requires understanding Temporal concepts

### Risks

- Temporal Server availability is critical
- Need to handle workflow versioning carefully

## References

- [Temporal Documentation](https://docs.temporal.io/)
- [Temporal Go SDK](https://github.com/temporalio/sdk-go)
- [Workers README](../../../internal/workers/README.md)
