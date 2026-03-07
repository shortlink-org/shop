# Use Case: Apply Promo Code

## Status

Planned only. No Go implementation exists yet.

This README preserves the original use-case contract so the specification is not lost after removing the dead Rust branch.

---

## Summary

**Input:** cart (or cart id) + promo code.  
**Output:** applied result with discount amount, or validation error.

Caller sends the current cart and a promo code string; Pricer validates the code, applies promo rules, and returns the resulting discount or an error if the code is invalid, expired, or not applicable.

---

## Expected Input

- Cart with items and pricing context
- Promo code string
- Optional context:
  - customer segment
  - current date/time
  - cart subtotal

---

## Expected Output

**Success**

- `applied: true`
- `discount_amount`
- optional `promo_id`, `promo_name`, `message`

**Validation failure**

- `applied: false`
- `error_code`
- `message`

---

## Suggested Flow

1. Receive cart + promo code.
2. Resolve promo definition.
3. Validate code and business constraints.
4. Evaluate promo discount.
5. Return applied result or validation error.

---

## Scenarios

See [tests/bdd/features/promo_code.feature](/Users/user/myprojects/shortlink/shop/pricer/tests/bdd/features/promo_code.feature).
