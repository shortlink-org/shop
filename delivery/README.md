# Delivery Service

<img width='200' height='200' src="./docs/public/logo.svg" alt="Delivery Service Logo">

> [!NOTE]
> Service for managing delivery operations - package lifecycle, courier management, and order dispatching.

## Getting started

We use Makefile for build and deploy.

```bash
$> make help # show help message with all commands and targets
```

## ADR

- **Common**:
  - [ADR-0001](./docs/ADR/decisions/0001-init.md) - Init project
  - [ADR-0002](./docs/ADR/decisions/0002-c4-system.md) - C4 system
- **Domain**:
  - [ADR-0003](./docs/ADR/decisions/0003-domain-structure.md) - Domain structure (model + services)
  - [ADR-0004](./docs/ADR/decisions/0004-dispatching-and-geolocation.md) - Dispatching and geolocation

## Use Cases

Package lifecycle (AcceptOrder, AssignOrder, PickUpOrder, DeliverOrder) and event semantics are described in [Use Cases Overview](./src/usecases/README.md). Quick index:

- [UC-1](./src/usecases/package/command/accept_order/README.md) Accept order from OMS
- [UC-2](./src/usecases/package/command/assign_order/README.md) Assign order to courier
- [UC-3](./src/usecases/package/command/deliver_order/README.md) Deliver order
- [UC-4](./src/usecases/courier/command/update_location/README.md) Update courier location
- [UC-5](./src/usecases/courier/command/register/README.md) Register courier
- [UC-6](./src/usecases/package/query/get_pool/README.md) Get package pool
- [UC-7](./src/usecases/courier/query/get_pool/README.md) Get courier pool

## Domain

The domain layer follows DDD principles with clear separation:

| Layer | Description |
|-------|-------------|
| `model/` | Aggregates, Value Objects, Proto definitions |
| `services/` | Domain services (pure business logic) |

Key components:

- **Package Aggregate** - State machine for package lifecycle
- **Courier Aggregate** - Status and capacity management
- **DispatchService** - Courier selection algorithm (Haversine)
- **AssignmentValidationService** - Business rules validation

See [domain/README.md](./src/domain/README.md) for details.

**Data model:** Package and Courier entities, status enums (PackageStatus, CourierStatus), and event types align with the ERD and proto in [docs/services-plan.md](../docs/services-plan.md) (section 4, Delivery). Domain types: [package/entity.rs](./src/domain/model/package/entity.rs), [courier/entity.rs](./src/domain/model/courier/entity.rs).

## Docs

- [Domain Models](./src/domain/README.md)
- [Use Cases Overview](./usecases/README.md)
