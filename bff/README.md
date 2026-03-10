# ShortLink Shop BFF

GraphQL Federation Gateway powered by [WunderGraph Cosmo Router](https://cosmo-docs.wundergraph.com/).

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Cosmo Router                             │
│                         (BFF :9991)                             │
└───────────────┬─────────────────┬─────────────────┬─────────────┘
                │                 │                 │
                ▼                 ▼                 ▼
        ┌───────────────┐ ┌───────────────┐ ┌───────────────┐
        │ carts-subgraph│ │ admin-subgraph│ │   countries   │
        │ (Go/Connect)  │ │ (Go/Connect)  │ │  (external)   │
        │    :4011      │ │    :4012      │ │               │
        └───────┬───────┘ └───────┬───────┘ └───────────────┘
                │                 │
                ▼                 ▼
        ┌───────────────┐ ┌───────────────┐
        │   OMS gRPC    │ │ Django Admin  │
        │    :50051     │ │    :8000      │
        └───────────────┘ └───────────────┘
```

## Subgraphs

| Subgraph   | Port | Technology | Description                        |
|------------|------|------------|------------------------------------|
| carts      | 4011 | Go/Connect | gRPC adapter for cart and orders   |
| admin      | 4012 | Go/Connect | REST adapter for goods             |
| countries  | -    | External   | External GraphQL API               |

## Prerequisites

- Node.js 18+
- pnpm
- Cosmo CLI (`wgc`, pinned to `0.109.0` in `package.json`)
- All subgraphs running

## Installation

```bash
pnpm install
```

## Usage

### 1. Start subgraphs

```bash
# Terminal 1: OMS gRPC service
cd ../oms && make run

# Terminal 2: Carts subgraph
cd ../oms-graphql && go run ./cmd/service

# Terminal 3: Admin subgraph (goods)
cd ../admin-graphql && go run ./cmd/service

# Terminal 4: Django Admin
cd ../admin && uv run python src/manage.py runserver 8000
```

### 2. Compose the federated schema

```bash
pnpm run compose
```

This generates `router-config.json` from `graph-local.yaml` for local development.

### 3. Download and run the router

```bash
# Download Cosmo Router binary
pnpm run download:router

# Start the router
./router --config router.yaml
```

The GraphQL endpoint will be available at `http://localhost:9991/graphql`.

## Docker

```bash
# Build image
pnpm run docker:build

# Run container
pnpm run docker:run
```

## Generating router-config.json for Deployment

The `router-config.json` file contains the composed federated schema and is required for the router to start.

### Option 1: Introspection (requires running subgraphs)

```bash
# Start all subgraphs first, then:
pnpm run compose
```

### Option 2: Schema Files (offline composition)

Use `graph-static.yaml` for deployment builds and `graph-local.yaml` for local development.
For grpc-mapped subgraphs, use `dns:///host:port` routing URLs:

```yaml
version: 1
subgraphs:
  - name: carts
    routing_url: dns:///localhost:4011
    grpc:
      schema_file: ../oms-graphql/pkg/graph/schema.graphql
      proto_file: ../oms-graphql/pkg/proto/service/v1/service.proto
      mapping_file: ../oms-graphql/pkg/proto/service/v1/mapping.json
  - name: admin
    routing_url: dns:///localhost:4012
    grpc:
      schema_file: ../admin-graphql/pkg/graph/schema.graphql
      proto_file: ../admin-graphql/pkg/proto/service/v1/service.proto
      mapping_file: ../admin-graphql/pkg/proto/service/v1/mapping.json
```

### CI/CD Integration

In your CI pipeline, compose the schema before building the Docker image:

```yaml
build:bff:
  script:
    - cd bff
    - npm install -g wgc@0.109.0
    - wgc router compose -i graph-static.yaml -o router-config.json
    - docker build -f ops/dockerfile/Dockerfile -t bff:latest .
```

**Note:** The `router-config.json` must be generated and included in the Docker image. It is copied during the Docker build process.

## Configuration

### graph-local.yaml

Defines local-development subgraphs for schema composition:

```yaml
version: 1
subgraphs:
  - name: carts
    routing_url: dns:///localhost:4011
    grpc:
      schema_file: ../oms-graphql/pkg/graph/schema.graphql
      proto_file: ../oms-graphql/pkg/proto/service/v1/service.proto
      mapping_file: ../oms-graphql/pkg/proto/service/v1/mapping.json
  - name: admin
    routing_url: dns:///localhost:4012
    grpc:
      schema_file: ../admin-graphql/pkg/graph/schema.graphql
      proto_file: ../admin-graphql/pkg/proto/service/v1/service.proto
      mapping_file: ../admin-graphql/pkg/proto/service/v1/mapping.json
  - name: countries
    routing_url: https://countries.trevorblades.com/
    introspection:
      url: https://countries.trevorblades.com/
```

### graph-static.yaml

Defines deployment-safe subgraphs baked into the Docker image during build:

```yaml
version: 1
subgraphs:
  - name: carts
    routing_url: dns:///shortlink-shop-oms-graphql.shortlink-shop.svc.cluster.local:4011
    grpc:
      schema_file: subgraphs/carts/schema.graphql
      proto_file: subgraphs/carts/service.proto
      mapping_file: subgraphs/carts/mapping.json
  - name: admin
    routing_url: dns:///shortlink-shop-admin-graphql.shortlink-shop.svc.cluster.local:4012
    grpc:
      schema_file: subgraphs/admin/schema.graphql
      proto_file: subgraphs/admin/service.proto
      mapping_file: subgraphs/admin/mapping.json
  - name: countries
    routing_url: https://countries.trevorblades.com/
    schema:
      file: schemas/countries.graphql
```

### router.yaml

Router runtime configuration (CORS, logging, metrics, etc.).

### Runtime Config

The router reads runtime settings from `router.yaml`.
Subgraph routing URLs are baked into `router-config.json` during image build from `graph-static.yaml`.
For grpc-mapped subgraphs, `routing_url` must use `dns:///host:port`, not `http://host:port`.

## Troubleshooting

### "Failed to fetch from Subgraph 'admin'" or "Failed to fetch from Subgraph 'carts'"

The BFF cannot reach the gRPC subgraphs. In Kubernetes the router expects these **Service** names in the **shortlink-shop** namespace:

| Subgraph | Expected Service name              | Port |
|----------|------------------------------------|------|
| admin    | `shortlink-shop-admin-graphql`      | 4012 |
| carts    | `shortlink-shop-oms-graphql`       | 4011 |

**Checks:**

1. **Deployments** – Ensure `admin-graphql` and `oms-graphql` are deployed. Helm release names should be `shortlink-shop-admin-graphql` and `shortlink-shop-oms-graphql` so that the created Services match the names above (if your chart uses `Release.Name` for the Service name).

2. **Pods** – Confirm subgraph pods are Running and Ready:
   ```bash
   kubectl get pods -n shortlink-shop -l app.kubernetes.io/name=admin-graphql
   kubectl get pods -n shortlink-shop -l app.kubernetes.io/name=oms-graphql
   ```

3. **Services** – Verify Services exist and target the correct port:
   ```bash
   kubectl get svc -n shortlink-shop shortlink-shop-admin-graphql shortlink-shop-oms-graphql
   ```

4. **Different names** – Update `graph-static.yaml`, rebuild `router-config.json` via Docker image build, and redeploy BFF.

## Example Queries

### Get goods

```graphql
query {
  goods {
    count
    results {
      id
      name
      price
    }
  }
}
```

### Get cart

```graphql
query {
  getCart {
    state {
      cartId
      items {
        goodId
        quantity
      }
    }
  }
}
```

### Get countries

```graphql
query {
  countries {
    code
    name
    emoji
  }
}
```

## Learn More

- [Cosmo Router Documentation](https://cosmo-docs.wundergraph.com/router/intro)
- [Cosmo CLI Reference](https://cosmo-docs.wundergraph.com/cli/intro)
- [GraphQL Federation](https://www.apollographql.com/docs/federation/)
