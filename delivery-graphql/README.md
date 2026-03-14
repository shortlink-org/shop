# Delivery GraphQL Subgraph

`delivery-graphql` is a Go/Connect federation subgraph for the Delivery service. It exposes the Cosmo subgraph contract over gRPC on port `4013` and translates those requests to the Delivery backend gRPC API.

## Architecture

```text
admin-ui / ui -> BFF (Cosmo Router) -> delivery-graphql (Go/Connect) -> Delivery service (gRPC)
```

## Running

```bash
# Generate/update Go stubs
buf generate

# Start the subgraph
go run ./cmd/service

# Run tests
go test ./...
```

## Environment variables

| Variable | Description | Default |
| --- | --- | --- |
| `LISTEN_ADDR` | Connect/gRPC listen address | `0.0.0.0:4013` |
| `DELIVERY_GRPC_URL` | Delivery backend gRPC target | `localhost:50052` |

## Subgraph contract

The federated contract is defined in:

- `pkg/graph/schema.graphql`
- `pkg/proto/service/v1/service.proto`
- `pkg/proto/service/v1/mapping.json`

The currently exposed query fields are:

- `randomAddress`
- `deliveryTracking(orderId: String!)`

Delivery tracking refreshes are now expected to come from repeated snapshot queries via the BFF rather than a GraphQL subscription endpoint in this service.

## References

- Migration notes: `docs/MIGRATION-TO-CONNECT.md`
- Delivery backend proto: `delivery/src/infrastructure/rpc/delivery.proto`
