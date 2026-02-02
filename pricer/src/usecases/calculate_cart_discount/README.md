# Use Case: Calculate Cart Discount

## Summary

**Input:** list of cart items (structure from buf registry).  
**Output:** discount amount for the cart.

Caller (OMS, BFF) sends the cart contents; Pricer evaluates discount policies (OPA/Rego) and returns the total discount amount.

---

## Input: Cart Items

Use the cart model from **buf registry** so Pricer stays aligned with OMS:

- **Module:** `buf.build/shortlink-org/shop-oms`
- **Messages:** `CartState` and `CartItem` from `infrastructure.rpc.cart.v1.model.v1`

From OMS cart model:

```protobuf
// CartItem (OMS)
message CartItem {
  string good_id = 1;
  int32  quantity = 2;
}

// CartState (OMS)
message CartState {
  string   cart_id     = 1;
  string   customer_id = 2;
  repeated CartItem items = 3;
  ...
}
```

For discount calculation we also need **price** and **brand** per item (OPA policies use them). Options:

1. **Pricer-specific request** — e.g. `CalculateCartDiscountRequest` with items that have `good_id`, `quantity`, `price`, `brand` (BFF/OMS fills prices from catalog).
2. **Reuse OMS types + extension** — request carries `CartState` and a separate map or list of `(good_id, price, brand)` from catalog.

Recommendation: define a small Pricer proto that references or embeds OMS cart IDs and adds pricing context (`price`, `brand`) so the use case has a single, clear input.

---

## Output: Discount Amount

- **Total discount** — one decimal amount (e.g. in minor units or as a string for precision).
- Optionally: **per-item breakdown** or **policy names** that were applied (for transparency).

Example response shape:

```text
discount_amount: "150.00"   # total discount
currency: "USD"             # optional
policies_applied: ["apple_samsung_discount", "general_discount"]
```

---

## Flow

1. **Receive** list of cart items (with good_id, quantity, price, brand).
2. **Build** OPA input from items + customer_id (and any extra params).
3. **Evaluate** discount policy (OPA/Rego) → raw discount value.
4. **Return** discount amount (and optionally applied policies).

---

## Dependencies

- **Port:** `DiscountPolicyEvaluator` (evaluate cart + params → discount).
- **Domain:** Cart/pricing context (value objects or proto-generated structs).
- **External:** Input shape compatible with buf `shop-oms` cart model; Pricer adds pricing fields where needed.

---

## BDD Scenarios

See [tests/bdd/features/quantity_discount.feature](../../../tests/bdd/features/quantity_discount.feature):

- **Buy 2, get 3rd free** — 3 items → discount = 1 unit price; 2 items → 0; 6 items → 2 units free.
- **Buy 3, get 4th with 25% off** — 4 items → discount = 25% of 1 unit; 3 items → 0; 8 items → 2 units at 25% off.
