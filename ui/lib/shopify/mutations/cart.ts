import cartFragment from '../fragments/cart';

export const addToCartMutation = /* GraphQL */ `
  mutation AddToCart($addRequest: ItemRequest!) {
    addItem(addRequest: $addRequest)
  }
`;

export const createCartMutation = /* GraphQL */ `
  mutation AddToCart($addRequest: ItemRequest!) {
    addItem(addRequest: $addRequest)
  }
`;

export const editCartItemsMutation = /* GraphQL */ `
  mutation AddToCart($addRequest: ItemRequest!) {
    addItem(addRequest: $addRequest)
  }
`;

export const removeFromCartMutation = /* GraphQL */ `
  mutation RemoveFromCart($removeRequest: ItemRequest!) {
    removeItem(removeRequest: $removeRequest)
  }
`;

export const resetCartMutation = /* GraphQL */ `
  mutation ResetCart($customerId: CartRequest!) {
    resetCart(customerId: $customerId)
  }
`;

export const checkoutMutation = /* GraphQL */ `
  mutation Checkout($input: CheckoutInput!) {
    checkout(input: $input) {
      orderId
    }
  }
`;
