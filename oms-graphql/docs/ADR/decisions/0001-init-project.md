# 1. Init project

Date: 2024-09-02

## Status

Superseded

## Context

I am working on an OMS GraphQL API Bridge to translate my gRPC API into a GraphQL interface for public API access. 
This is necessary to expose the gRPC functionality to a broader audience that utilizes GraphQL.

## Decision

At the time, we chose `https://tailcall.run/` as the bridge implementation.
That decision has since been replaced by a Go-based Cosmo Connect gRPC subgraph implementation that talks to `oms` directly over gRPC.

### Alternatives

#### The Guild's GraphQL Mesh

Another alternative considered was `https://the-guild.dev/graphql/mesh/docs/handlers/grpc#use-reflection-instead-of-proto-files`, 
which offers capabilities for integrating gRPC with GraphQL, including the use of reflection instead of proto files. 
However, this tool was not selected due to a known issue (`Invalid value used as weak map key`) that has not yet been resolved, 
making the product unreliable for our use case.

## Consequences

This document is kept only as historical context.
The active implementation no longer depends on Tailcall and uses `buf`-generated protobuf code plus a Go Connect service instead.
