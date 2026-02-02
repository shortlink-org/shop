# BDD tests (Gherkin)

Acceptance tests in Gherkin format. Run with **cucumber-rs** (see ROADMAP Phase 3).

## Features

| Feature | Scenarios | Use case |
|---------|-----------|----------|
| [quantity_discount.feature](./features/quantity_discount.feature) | Buy 2 get 3rd free; Buy 3 get 4th with 25% off | Calculate Cart Discount |
| [promo_code.feature](./features/promo_code.feature) | Promo -10%; Promo -10% from 3000; invalid/expired | Apply Promo Code |

## Test cases (summary)

1. **2 покупаем, третий в подарок** — `quantity_discount.feature`: 3 items → discount 100 (1 free), 2 items → 0, 6 items → 200 (2 free).
2. **3 покупаем, на 4-й скидка 25%** — `quantity_discount.feature`: 4 items → discount 50 (25% of 200), 3 items → 0, 8 items → 100.
3. **Промокод -10%** — `promo_code.feature`: SAVE10, no minimum → 10% off (e.g. 1000 → 100 discount).
4. **Промокод -10% от суммы покупки от 3000** — `promo_code.feature`: BIG10, min 3000 → applied when subtotal ≥ 3000 (300 discount), rejected when &lt; 3000 (MIN_ORDER_NOT_MET).

Step definitions and wiring to use cases are to be implemented in Phase 3 (ROADMAP).
