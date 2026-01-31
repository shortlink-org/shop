export const getGoodQuery = /* GraphQL */ `
  query GetGood($id: Int!) {
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

