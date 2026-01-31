# Admin GraphQL Subgraph

GraphQL subgraph for the Django Admin API, converting REST endpoints to GraphQL using [Tailcall](https://tailcall.run/).

## Architecture

This service acts as a GraphQL layer over the Django Admin REST API:

```
Cosmo Router → admin-graphql (Tailcall) → Django Admin API
     :9991           :8101                      :8000
```

## Supported Resources

- **Goods** - Product catalog
- **Offices** - Pickup locations with geolocation

## Prerequisites

- Node.js 18+
- pnpm (recommended) or npm
- Django Admin service running on port 8000

## Installation

```bash
pnpm install
```

## Running

```bash
# Start the service
pnpm start

# Development mode with watch
pnpm dev
```

The GraphQL endpoint will be available at `http://localhost:8101/graphql`.

## Configuration

The Tailcall configuration is in `config/admin.graphql`.

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `ADMIN_API_URL` | `http://127.0.0.1:8000` | Django Admin API base URL |

## GraphQL Schema

### Queries

#### Goods
- `goods(page: Int): GoodsList` - Get paginated list of goods
- `good(id: Int!): Good` - Get a single good by ID

#### Offices
- `offices`: [Office!]!` - Get all offices
- `office(id: Int!): Office` - Get a single office by ID

### Types

```graphql
type Good {
  id: Int!
  name: String!
  price: String!
  description: String!
  created_at: String!
  updated_at: String!
}

type Office {
  id: Int!
  name: String!
  address: String!
  latitude: Float
  longitude: Float
  opening_time: String!
  closing_time: String!
  working_days: String!
  phone: String
  email: String
  is_active: Boolean!
  created_at: String!
  updated_at: String!
}
```

## Example Queries

```graphql
# Get all goods
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

# Get single good
query {
  good(id: 1) {
    id
    name
    price
    description
  }
}

# Get all offices
query {
  offices {
    id
    name
    address
    latitude
    longitude
    is_active
  }
}

# Get single office
query {
  office(id: 1) {
    id
    name
    address
    opening_time
    closing_time
    working_days
  }
}
```
