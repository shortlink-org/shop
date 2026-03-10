# Migration: Delivery subgraph to Cosmo Connect (Go)

## Goal

Bring the delivery subgraph to the same pattern as `oms-graphql` and `admin-graphql`: a **Go/Connect** service that implements the Federation subgraph protocol over gRPC and calls the Delivery service via generated gRPC clients, instead of Tailcall (Node) with runtime JSON→gRPC mapping.

## Current state

- **delivery-graphql** (this repo): Tailcall, Node, port 8080, HTTP GraphQL with `enableFederation: true`. BFF composes it via introspection (graph.yaml) or static schema (when added to graph-static).
- **Delivery service**: Rust, gRPC API in `delivery/src/infrastructure/rpc/delivery.proto`.

## Target state

- New **delivery-graphql** implementation in Go (or replace this repo):
  - **Buf** for proto: subgraph contract proto (Cosmo Connect format, e.g. `shop.delivery.v1` with `QueryCouriers`, `QueryCourier`, `QueryCourierDeliveries`, `MutationRegisterCourier`, …) + dependency on Delivery proto (or copy under `pkg/proto/delivery`).
  - **Connect** handlers: each handler receives the subgraph request, maps to Delivery gRPC request, calls Delivery client, maps response back.
  - **Federation schema**: same GraphQL schema as today (`config/delivery.graphql` or `bff/schemas/delivery.graphql`), plus `mapping.json` and generated `service.proto` for the subgraph (see `oms-graphql/pkg/proto/service/v1/`).
  - **Port**: e.g. 4013 (gRPC), same as carts 4011, admin 4012.
- BFF **graph.yaml** / **graph-static.yaml**: delivery subgraph with `routing_url: dns:///...:4013`, `grpc: { schema_file, proto_file, mapping_file }` (no introspection).

## Steps (outline)

1. **Scaffold Go module** (e.g. in this repo or new `delivery-graphql-go`):
   - `go.mod`, `buf.yaml`, `buf.gen.yaml` (Connect + gRPC Go, subgraph proto).
   - Depend on Delivery proto: buf dependency on Delivery repo/module, or vendor `delivery.proto` and generate Go client.

2. **Define subgraph proto** (Cosmo Connect contract):
   - One RPC per GraphQL root field: `QueryCouriers`, `QueryCourier`, `QueryCourierDeliveries`, `MutationRegisterCourier`, `MutationActivateCourier`, etc. Request/response messages mirror the GraphQL types (see oms-graphql `service.proto` and `mapping.json`).

3. **GraphQL schema + mapping**:
   - Copy/adapt `config/delivery.graphql` as the subgraph SDL.
   - Run Cosmo codegen (or by hand) to get `mapping.json` from schema + subgraph proto.

4. **Implement handlers**:
   - For each RPC: read Connect request, map to Delivery gRPC request, call Delivery client (with traceparent/trace-id in metadata), map response to Connect response.
   - Propagate **traceparent** and **trace-id** from incoming headers to outgoing gRPC metadata (same as oms-graphql).

5. **Deploy**:
   - Helm chart: same pattern as oms-graphql (name e.g. `shortlink-shop-delivery-graphql`, port 4013, env `DELIVERY_GRPC_URL`).
   - BFF: add delivery to `graph-static.yaml` with `dns:///shortlink-shop-delivery-graphql.shortlink-shop.svc.cluster.local:4013`, include schema/proto/mapping in BFF image (e.g. `subgraphs/delivery/`).

6. **Retire Tailcall**:
   - Remove Tailcall config and Node dependency; optionally keep this repo as Go-only or merge into a single “graphql-bridges” repo.

## References

- oms-graphql: `pkg/service/service.go` (handlers, metadata propagation), `pkg/proto/service/v1/service.proto`, `pkg/graph/schema.graphql`, `pkg/proto/service/v1/mapping.json`.
- admin-graphql: same structure, HTTP client to Django instead of gRPC.
- BFF: `bff/graph-static.yaml`, `bff/graph.yaml`, `bff/router.yaml` (header propagation).
- Delivery gRPC: `delivery/src/infrastructure/rpc/delivery.proto`.
