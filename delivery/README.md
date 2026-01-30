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

## Use Cases

- [UC-1](./usecases/accept_order/README.md) Accept order from OMS
- [UC-2](./usecases/assign_order/README.md) Assign order to courier
- [UC-3](./usecases/deliver_order/README.md) Deliver order
- [UC-4](./usecases/update_courier_location/README.md) Update courier location
- [UC-5](./usecases/register_courier/README.md) Register courier
- [UC-6](./usecases/get_package_pool/README.md) Get package pool
- [UC-7](./usecases/get_courier_pool/README.md) Get courier pool

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

## Docs

- [Domain Models](./src/domain/README.md)
- [Use Cases Overview](./usecases/README.md)
