# Delivery GraphQL Subgraph

GraphQL subgraph for the Delivery service. It currently proxies the gRPC API to GraphQL via [Tailcall](https://tailcall.run/) (Node). The target approach is the same pattern as **oms-graphql** and **admin-graphql**: a Go/Connect subgraph with typed calls to the Delivery gRPC API. Migration plan: [docs/MIGRATION-TO-CONNECT.md](./docs/MIGRATION-TO-CONNECT.md).

## Architecture

```
admin-ui → BFF (Cosmo Router) → delivery-graphql → Delivery Service (gRPC)
```

## Running

```bash
# Install dependencies
pnpm install

# Start (port 8080)
pnpm start

# Development with hot-reload
pnpm dev

# Validate configuration
pnpm check
```

## Environment variables

| Variable           | Description              | Default                    |
|--------------------|--------------------------|----------------------------|
| `DELIVERY_GRPC_URL` | Delivery gRPC service URL | `http://localhost:50051`   |

## GraphQL API

### Queries

- `couriers(filter, pagination)` — list couriers with filtering and pagination
- `courier(id, includeLocation)` — get a courier by ID
- `courierDeliveries(courierId, limit)` — courier delivery history

### Mutations

- `registerCourier(input)` — register a courier
- `activateCourier(id)` — activate courier
- `deactivateCourier(id, reason)` — deactivate courier
- `archiveCourier(id, reason)` — archive courier
- `updateCourierContact(id, input)` — update contact info
- `updateCourierSchedule(id, input)` — update schedule
- `changeCourierTransport(id, transportType)` — change transport type

## References

- [Tailcall Documentation](https://tailcall.run/docs/)
- Delivery Service proto: `delivery/src/infrastructure/rpc/delivery.proto`
