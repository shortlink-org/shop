export const getCartQuery = /* GraphQL */ `
  query GetCart {
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
`;
