# Use Case: Calculate Total

## Summary

**Input:** `domain.Cart` plus `discountParams` and `taxParams`.  
**Output:** `domain.CartTotal` with total discount, total tax, final price, and applied policy names.

This is the active Go use case used by Pricer to evaluate OPA-based discount and tax policies for a cart and return the final pricing result.

---

## Input

The command receives:

- `Cart`
  - `customerId`
  - `items[]`
    - `productId`
    - `quantity`
    - `price`
- `discountParams`
- `taxParams`

See:

- [command.go](/Users/user/myprojects/shortlink/shop/pricer/internal/usecases/cart/command/calculate_total/command.go)
- [cart.go](/Users/user/myprojects/shortlink/shop/pricer/internal/domain/cart.go)

---

## Output

The handler returns `domain.CartTotal`:

- `totalTax`
- `totalDiscount`
- `finalPrice`
- `policies`

See [cart_total.go](/Users/user/myprojects/shortlink/shop/pricer/internal/domain/cart_total.go).

---

## Flow

1. Receive cart and policy parameters.
2. Evaluate discount policy.
3. Evaluate tax policy.
4. Calculate subtotal from cart items.
5. Cap discount by subtotal to avoid negative итог.
6. Return `finalPrice = subtotal - totalDiscount + totalTax`.

Implementation: [handler.go](/Users/user/myprojects/shortlink/shop/pricer/internal/usecases/cart/command/calculate_total/handler.go)

---

## Related Specs

- Quantity and combination discount scenarios are described in [tests/bdd/features/quantity_discount.feature](/Users/user/myprojects/shortlink/shop/pricer/tests/bdd/features/quantity_discount.feature).
- Promo-code behavior is documented separately as planned functionality in [../apply_promo_code/README.md](/Users/user/myprojects/shortlink/shop/pricer/internal/usecases/cart/command/apply_promo_code/README.md).
