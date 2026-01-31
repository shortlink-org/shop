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
        │  (Tailcall)   │ │  (Tailcall)   │ │  (external)   │
        │    :8100      │ │    :8101      │ │               │
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
| carts      | 8100 | Tailcall   | gRPC → GraphQL for cart            |
| admin      | 8101 | Tailcall   | REST → GraphQL for goods, offices  |
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
cd ../oms-graphql && pnpm start

# Terminal 3: Admin subgraph (goods, offices)
cd ../admin-graphql && pnpm start

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
    routing_url: http://localhost:8100/graphql
    schema:
      file: ../oms-graphql/config/grpc.graphql
  - name: admin
    routing_url: http://localhost:8101/graphql
    schema:
      file: ../admin-graphql/config/admin.graphql
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
    routing_url: http://localhost:8100/graphql
    introspection:
      url: http://localhost:8100/graphql
  - name: admin
    routing_url: http://localhost:8101/graphql
    introspection:
      url: http://localhost:8101/graphql
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
| `CARTS_SUBGRAPH_URL`  | `http://localhost:8100/graphql`      | Carts subgraph URL         |
| `ADMIN_SUBGRAPH_URL`  | `http://localhost:8101/graphql`      | Admin subgraph URL         |
| `COUNTRIES_SUBGRAPH_URL` | `https://countries.trevorblades.com/` | Countries subgraph URL |
| `LOG_LEVEL`           | `info`                               | Log level                  |

## Example Queries

### Get goods

```graphql
query {
  goods(page: 1) {
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
  getCart(customerId: { customerId: "user-123" }) {
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

### Get offices

```graphql
query {
  offices {
    results {
      id
      name
      address
      latitude
      longitude
      is_active
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
