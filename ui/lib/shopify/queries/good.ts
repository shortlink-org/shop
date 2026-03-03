export const getGoodQuery = /* GraphQL */ `
  query GetGood($id: String!) {
    good(id: $id) {
      id
      name
      price
      description
      created_at
      updated_at
    }
  }
`;

export const getGoodsQuery = /* GraphQL */ `
  # NOTE: BFF goods endpoint currently supports pagination only.
  # Search and sort are applied client-side in lib/shopify/api/good.ts.
  query GetGoods($page: Int) {
    goods(page: $page) {
      count
      next
      previous
      results {
        id
        name
        price
        description
        created_at
        updated_at
      }
    }
  }
`;

export const getGoodRecommendationsQuery = /* GraphQL */ `
  query GetGoodRecommendations($page: Int) {
    goods(page: $page) {
      count
      next
      previous
      results {
        id
        name
        price
        description
      }
    }
  }
`;
