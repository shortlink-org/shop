# Admin GraphQL Subgraph

Cosmo Connect gRPC subgraph that adapts Django Admin REST endpoints to the federated shop graph.

## Architecture

This service acts as a typed adapter over the Django Admin REST API:

```
Cosmo Router → admin-graphql (Go/Connect) → Django Admin API
     :9991              :4012                       :8000
```

## Supported Resources

- **Goods** - Product catalog

## Prerequisites

- Go 1.25+
- buf
- Django Admin service running on port 8000

## Running

```bash
# Generate protobuf/Connect code
buf generate

# Start the service
go run ./cmd/service
```

The subgraph endpoint will be available at `http://localhost:4012`.

## Configuration

The federated schema is in `pkg/graph/schema.graphql`.

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `ADMIN_API_URL` | `http://127.0.0.1:8000` | Django Admin API base URL |
| `LISTEN_ADDR` | `0.0.0.0:4012` | Connect/gRPC listen address |

## GraphQL Schema

### Queries

- `goods: GoodsList` - Get paginated list of goods
- `good(id: String!): Good` - Get a single good by ID (UUID)

### Types

```graphql
type Good {
  id: String!
  name: String!
  price: String!
  description: String!
  created_at: String!
  updated_at: String!
}
```

## Example Queries

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

query {
  good(id: "550e8400-e29b-41d4-a716-446655440000") {
    id
    name
    price
    description
  }
}
```
