# Order management system (OMS) service

<img width='200' height='200' src="./docs/public/logo.svg" alt="OMS Service Logo">

> [!NOTE]
> Service for work with carts, orders.

## Getting started

We use Makefile for build and deploy.

```bash
$> make help # show help message with all commands and targets
```

## ADR

- **Common**:
  - [ADR-0001](./docs/ADR/decisions/0001-init.md) - Init project
  - [ADR-0002](./docs/ADR/decisions/0002-c4-system.md) - C4 system
- **Infrastructure**:
  - [ADR-0003](./docs/ADR/decisions/0003-temporal.md) - Temporal for workflow orchestration

## Use Cases

- [UC-1](internal/usecases/cart/README.md) Cart workflows
- [UC-2](internal/usecases/order/README.md) Order workflows

## Architecture

### Domain Layer

See [domain/README.md](./internal/domain/README.md) for domain structure:

- **Aggregates**: Cart, Order
- **Domain Services**: StockCartService, CartValidationService
- **Value Objects**: Address, Location, Weight

### Temporal Workers

See [workers/README.md](./internal/workers/README.md) for workflow documentation:

- **Cart Workflow** - manages cart state via signals (ADD, REMOVE, RESET)
- **Order Workflow** - manages order lifecycle (CREATE, COMPLETE, CANCEL)

## Docs

- [Domain Layer](./internal/domain/README.md)
- [Domain Services](./internal/domain/services/README.md)
- [Temporal Workers](./internal/workers/README.md)
