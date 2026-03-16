export const addToCartMutation = /* GraphQL */ `
  mutation AddToCart($addRequest: ItemRequest!) {
    addItem(addRequest: $addRequest) {
      _
    }
  }
`;

export const createCartMutation = /* GraphQL */ `
  mutation AddToCart($addRequest: ItemRequest!) {
    addItem(addRequest: $addRequest) {
      _
    }
  }
`;

export const editCartItemsMutation = /* GraphQL */ `
  mutation AddToCart($addRequest: ItemRequest!) {
    addItem(addRequest: $addRequest) {
      _
    }
  }
`;

export const removeFromCartMutation = /* GraphQL */ `
  mutation RemoveFromCart($removeRequest: ItemRequest!) {
    removeItem(removeRequest: $removeRequest) {
      _
    }
  }
`;

export const resetCartMutation = /* GraphQL */ `
  mutation ResetCart {
    resetCart {
      _
    }
  }
`;

export const checkoutMutation = /* GraphQL */ `
  mutation Checkout($input: CheckoutInput!) {
    checkout(input: $input) {
      orderId
    }
  }
`;
