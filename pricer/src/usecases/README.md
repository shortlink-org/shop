# Pricer Use Cases

Application use cases: orchestration of domain and infrastructure to fulfill a single request.

| Use Case | Input | Output | README |
|----------|--------|--------|--------|
| **Calculate Cart Discount** | List of cart items (buf: `shop-oms` CartState/CartItem) + pricing context | Discount amount | [calculate_cart_discount](./calculate_cart_discount/README.md) |
| **Apply Promo Code** | Cart + promo code | Applied result: discount amount or validation error | [apply_promo_code](./apply_promo_code/README.md) |

## Cart structure (buf registry)

Cart items and cart state come from **buf.build/shortlink-org/shop-oms**:

- Package: `infrastructure.rpc.cart.v1.model.v1`
- Messages: `CartItem` (good_id, quantity), `CartState` (cart_id, customer_id, items, timestamps)

Pricer adds pricing context (price, brand) where needed for discount/tax policies. See each use case README for request/response shape.
