# Migration: Delivery subgraph to Cosmo Connect (Go)

## Status

The migration is complete. `delivery-graphql` now follows the same native Cosmo Connect/gRPC subgraph pattern as `oms-graphql` and `admin-graphql`.

## Current state

- `delivery-graphql` is a Go/Connect service exposing the federation contract on port `4013`.
- The federated contract lives in `pkg/graph/schema.graphql`, `pkg/proto/service/v1/service.proto`, and `pkg/proto/service/v1/mapping.json`.
- The BFF composes `delivery` as a grpc-mapped subgraph in `bff/graph.yaml`, `bff/graph-local.yaml`, and `bff/graph-static.yaml`.
- `bff/subgraphs/delivery/` is synchronized from `delivery-graphql` via `buf export` before router composition.

## Final architecture

- **Transport to BFF:** native Cosmo Connect/gRPC subgraph on `4013`
- **Transport to backend:** generated Delivery gRPC client
- **Code generation:** `buf generate` for Go stubs, `buf export` for BFF proto sync
- **Schema composition:** `wgc router compose` against SDL + proto + mapping artifacts

## Subscription strategy

The old GraphQL subscription path (`deliveryTrackingUpdated`) was intentionally not carried into the native subgraph contract.

- The current `service.proto` and `mapping.json` only expose query operations.
- UI clients should obtain updated tracking state through repeated `deliveryTracking(orderId)` snapshot queries.
- This keeps the subgraph transport aligned with the other native Cosmo Connect subgraphs and removes the in-process GraphQL runtime from `delivery-graphql`.

## Operational notes

1. Generate Go stubs in `delivery-graphql` with `buf generate`.
2. Sync delivery artifacts into the BFF with `pnpm run sync:delivery-subgraph` in `bff/`.
3. Compose the router config with `pnpm run compose:local` or `pnpm run compose:static`.

## References

- `oms-graphql/pkg/service/service.go`
- `oms-graphql/pkg/proto/service/v1/service.proto`
- `admin-graphql/pkg/service/service.go`
- `bff/graph-static.yaml`
- `bff/scripts/sync-delivery-subgraph.sh`
- `delivery/src/infrastructure/rpc/delivery.proto`
