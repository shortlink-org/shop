# language: en
# Promo code application: flat percent off and percent off with minimum order

Feature: Apply promo code
  As a customer
  I want to apply a promo code at checkout
  So that I get an extra discount when I have a valid code

  Scenario: Apply promo code with 10% off (no minimum)
    Given a cart with items:
      | good_id | quantity | unit_price | brand  |
      | good-A  | 2        | 500.00     | Acme   |
    And the cart subtotal is 1000.00
    And promo code "SAVE10" gives 10% off with no minimum order
    When I apply promo code "SAVE10" to the cart
    Then the promo is applied successfully
    And the discount from promo is 100.00
    And the total discount is 100.00

  Scenario: Apply promo code with 10% off — small cart (no minimum)
    Given a cart with items:
      | good_id | quantity | unit_price | brand  |
      | good-A  | 1        | 100.00     | Acme   |
    And the cart subtotal is 100.00
    And promo code "SAVE10" gives 10% off with no minimum order
    When I apply promo code "SAVE10" to the cart
    Then the promo is applied successfully
    And the discount from promo is 10.00

  Scenario: Apply promo code with 10% off on purchase from 3000 — cart meets minimum
    Given a cart with items:
      | good_id | quantity | unit_price | brand  |
      | good-A  | 2        | 1500.00    | Acme   |
    And the cart subtotal is 3000.00
    And promo code "BIG10" gives 10% off when order amount is at least 3000
    When I apply promo code "BIG10" to the cart
    Then the promo is applied successfully
    And the discount from promo is 300.00
    And the total discount is 300.00

  Scenario: Apply promo code with 10% off on purchase from 3000 — cart below minimum
    Given a cart with items:
      | good_id | quantity | unit_price | brand  |
      | good-A  | 2        | 1400.00    | Acme   |
    And the cart subtotal is 2800.00
    And promo code "BIG10" gives 10% off when order amount is at least 3000
    When I apply promo code "BIG10" to the cart
    Then the promo is not applied
    And the error code is "MIN_ORDER_NOT_MET"
    And the message indicates minimum order is 3000.00

  Scenario: Apply promo code with 10% off on purchase from 3000 — cart exactly 3000
    Given a cart with items:
      | good_id | quantity | unit_price | brand  |
      | good-A  | 1        | 3000.00    | Acme   |
    And the cart subtotal is 3000.00
    And promo code "BIG10" gives 10% off when order amount is at least 3000
    When I apply promo code "BIG10" to the cart
    Then the promo is applied successfully
    And the discount from promo is 300.00

  Scenario: Apply invalid promo code
    Given a cart with items:
      | good_id | quantity | unit_price | brand  |
      | good-A  | 1        | 100.00     | Acme   |
    When I apply promo code "INVALID" to the cart
    Then the promo is not applied
    And the error code is "INVALID_CODE"

  Scenario: Apply expired promo code
    Given a cart with items:
      | good_id | quantity | unit_price | brand  |
      | good-A  | 1        | 100.00     | Acme   |
    And promo code "EXPIRED10" is expired
    When I apply promo code "EXPIRED10" to the cart
    Then the promo is not applied
    And the error code is "EXPIRED"
