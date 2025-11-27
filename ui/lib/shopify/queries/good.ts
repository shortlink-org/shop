export const getGoodQuery = /* GraphQL */ `
  query Goods_goods_retrieve(
    $id: Int!
  ) {
    goods_goods_retrieve(id: $id) {
        created_at
        description
        id
        name
        price
        updated_at
    }
  }
`;

export const getGoodsQuery = /* GraphQL */ `
  query Goods_goods_retrieve {
    goods_goods_list {
      count
      next
      previous
    }
  }
`;

export const getGoodRecommendationsQuery = /* GraphQL */ `
  query Goods_goods_retrieve {
    goods_goods_list {
      count
      next
      previous
    }
  }
`;

