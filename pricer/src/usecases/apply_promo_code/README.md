# Use Case: Apply Promo Code

## Summary

**Input:** cart (or cart id) + promo code.  
**Output:** updated discount (or result of applying the promo): success + new discount amount, or validation error.

Caller sends the current cart and a promo code string; Pricer validates the code, applies promo rules (OPA or domain logic), and returns the resulting discount (or error if code invalid/expired/not applicable).

---

## Input

- **Cart** — same as in [Calculate Cart Discount](../calculate_cart_discount/README.md): list of items (from buf registry) with pricing context (good_id, quantity, price, brand), plus `customer_id`.
- **Promo code** — string (e.g. `"SUMMER10"`, `"WELCOME5"`).

Optional context for rules:

- Customer segment / registration date (for “new customers only”).
- Current date/time (for validity window).
- Cart subtotal (for “min order” rules).

---

## Output

**Success:**

- `applied: true`
- `discount_amount` — total discount after applying promo (may replace or stack with other discounts, depending on business rules).
- `promo_id` / `promo_name` — for display and idempotency.
- Optionally: `policies_applied`, `message` (e.g. “10% off applied”).

**Validation failure:**

- `applied: false`
- `error_code`: e.g. `INVALID_CODE`, `EXPIRED`, `MIN_ORDER_NOT_MET`, `NOT_APPLICABLE`.
- `message` — human-readable reason.

---

## Flow

1. **Receive** cart + promo code (and optional context).
2. **Resolve** promo code → promo definition (from store/config or OPA data).
3. **Validate** — code exists, not expired, conditions met (min order, customer segment, etc.).
4. **Compute** discount for this cart under this promo (OPA or domain service).
5. **Return** applied result (new discount amount) or validation error.

---

## Dependencies

- **Port:** `PromoCodeResolver` (code string → promo definition/rule).
- **Port:** `DiscountPolicyEvaluator` or dedicated `PromoDiscountEvaluator` (cart + promo → discount).
- **Domain:** Promo (value object or proto): code, type (percent/fixed), conditions, validity.
- **Storage/Config:** promo definitions — config file, DB, or OPA data; TBD in implementation.

---

## Relation to Calculate Cart Discount

- **Calculate Cart Discount** — “automatic” discounts only (e.g. by brand, category, quantity).
- **Apply Promo Code** — user-entered code; can stack with automatic discounts or override them (business rule).

Both can share the same OPA discount pipeline with different inputs (with/without `promo_code` and promo rules).

---

## BDD Scenarios

See [tests/bdd/features/promo_code.feature](../../../tests/bdd/features/promo_code.feature):

- **Promo -10% (no minimum)** — SAVE10: 10% off any order (e.g. subtotal 1000 → discount 100).
- **Promo -10% from 3000** — BIG10: 10% off when subtotal ≥ 3000; applied at 3000 (discount 300), rejected below 3000 (MIN_ORDER_NOT_MET).
- **Invalid / expired code** — INVALID_CODE, EXPIRED.
