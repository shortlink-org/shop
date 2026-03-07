# 3. Migrate To Cosmo Connect gRPC Subgraph

Date: 2026-03-08

## Status

Accepted

## Context

The original `oms-graphql` implementation used Tailcall to translate GraphQL operations into gRPC calls to `oms`.
That approach was fast to bootstrap, but it created a weak point in a high-value path:

- request payloads for gRPC methods were assembled from string templates at runtime
- the effective contract was split across GraphQL schema, Tailcall directives, proto files, and header scripts
- production failures were hard to reason about because errors surfaced as downstream parse failures instead of typed compile-time errors
- the router could not treat `oms-graphql` as a native gRPC subgraph in Cosmo Connect

This became a practical problem when cart operations started failing with:

`Failed to parse input according to type infrastructure.rpc.cart.v1.model.v1.GetRequest`

and the same class of error for `AddRequest`.

Direct calls to `oms` over gRPC succeeded, which confirmed that the instability was in the translation layer rather than in the backend RPC handlers.

## Decision

We will replace the Tailcall-based bridge with a Go-based Cosmo Connect gRPC subgraph.

The new implementation will:

- use `buf` as the source of truth for protobuf code generation
- use remote plugins from the Buf Schema Registry for generated Go, gRPC, and Connect code
- expose a typed Connect/gRPC service that matches the federated schema expected by Cosmo Router
- call `oms` directly through generated gRPC clients instead of runtime JSON templating
- enforce the `X-User-ID` contract in Go code and forward it to `oms` metadata explicitly

The federated GraphQL contract remains the public API surface for the router, but the translation from subgraph contract to backend gRPC contract is now implemented in typed Go code.

## Alternatives

### Keep Tailcall and patch the directives

This was rejected because it would preserve the same runtime-templating failure mode and keep request construction spread across templates and scripts.

### Route `bff` directly to the existing `oms` gRPC API

This was rejected as a short-term change because the current Cosmo setup composes subgraphs, not arbitrary existing backend gRPC services.
It still requires a dedicated federated gRPC subgraph contract and an adapter layer.

### Generate a new standalone adapter from GraphQL schema without keeping Go code in-repo

This was rejected because we need explicit control over header propagation, request mapping, deployment shape, observability, and future evolution of cart/order operations.

## Consequences

Positive:

- request/response mapping is now compile-time checked
- `oms-graphql` can be composed as a native Cosmo Connect gRPC subgraph
- protobuf and Connect code generation is reproducible through `buf` and BSR
- user identity propagation is explicit and testable
- failures are easier to localize to either the adapter or `oms`

Negative:

- the service is now a maintained Go codebase instead of a mostly declarative bridge
- mapping logic between federated schema and backend RPC schema must be owned in code
- generated proto artifacts must be kept in sync with both the federated contract and `oms` RPC definitions

## Consequences For Existing ADRs

This decision supersedes the implementation choice recorded in ADR-0001.
ADR-0001 remains as historical context for how the project started, but it is no longer the active implementation direction.
