## OMS GraphQL API Bridge

> [!NOTE]
> Cosmo Connect gRPC subgraph that adapts federated cart/order operations to the `oms` gRPC API.

### Getting started

Generate code from protobuf definitions with `buf` and remote plugins from the Buf Schema Registry:

```bash
buf generate
```

Run the subgraph locally:

```bash
go run ./cmd/service
```

### API

The service exposes:
- gRPC/Connect handler on `:4011`
- health endpoint at `/healthz`
- Prometheus metrics at `/metrics`

### ADR

- **Common**:
  - [ADR-0001](./docs/ADR/decisions/0001-init-project.md) - Init project
  - [ADR-0002](./docs/ADR/decisions/0002-c4-system.md) - C4 system
  - [ADR-0003](./docs/ADR/decisions/0003-migrate-to-cosmo-connect-grpc.md) - Migrate to Cosmo Connect gRPC subgraph
