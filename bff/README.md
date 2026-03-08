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
- Cosmo CLI (`wgc`)
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

This generates `router-config.json` from all subgraph schemas.

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

Update `graph.yaml` to use schema files instead of introspection:

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
    - npm install -g wgc
    - wgc router compose -i graph.yaml -o router-config.json
    - docker build -f ops/dockerfile/Dockerfile -t bff:latest .
```

**Note:** The `router-config.json` must be generated and included in the Docker image. It is copied during the Docker build process.

## Configuration

### graph.yaml

Defines subgraphs for schema composition:

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

### router.yaml

Router runtime configuration (CORS, logging, metrics, etc.).

### Environment Variables

| Variable              | Default                              | Description                |
|-----------------------|--------------------------------------|----------------------------|
| `CARTS_SUBGRAPH_URL`  | `http://localhost:4011`              | Carts subgraph runtime URL |
| `ADMIN_SUBGRAPH_URL`  | `http://localhost:4012`              | Admin subgraph runtime URL |
| `COUNTRIES_SUBGRAPH_URL` | `https://countries.trevorblades.com/` | Countries subgraph URL |
| `LOG_LEVEL`           | `info`                               | Log level                  |

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
