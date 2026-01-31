import cartFragment from '../fragments/cart';

export const getCartQuery = /* GraphQL */ `
  query GetCart($customerId: String!) {
    getCart(customerId: { customerId: $customerId }) {
      state {
        cartId
        customerId
        items {
          goodId
          quantity
        }
      }
    }
  }
`;
