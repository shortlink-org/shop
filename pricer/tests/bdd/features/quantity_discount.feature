# language: en
# Quantity-based discount rules (buy N, get M free or with discount)

Feature: Quantity discount
  As a customer
  I want discounts when I buy multiple units
  So that I benefit from "buy more, save more" offers

  Scenario: Buy 2, get 3rd free (2+1)
    Given a cart with items:
      | good_id | quantity | unit_price | brand  |
      | good-A  | 3        | 100.00     | Acme   |
    When I calculate the cart discount
    Then the discount amount should be 100.00
    And the discount is applied as "3rd item free" (1 free unit at 100.00)

  Scenario: Buy 2, get 3rd free — only 2 items in cart (no discount)
    Given a cart with items:
      | good_id | quantity | unit_price | brand  |
      | good-A  | 2        | 100.00     | Acme   |
    When I calculate the cart discount
    Then the discount amount should be 0.00

  Scenario: Buy 2, get 3rd free — 6 items (2 free units)
    Given a cart with items:
      | good_id | quantity | unit_price | brand  |
      | good-A  | 6        | 100.00     | Acme   |
    When I calculate the cart discount
    Then the discount amount should be 200.00
    And the discount is applied as "3rd item free" (2 free units)

  Scenario: Buy 3, get 4th with 25% off
    Given a cart with items:
      | good_id | quantity | unit_price | brand  |
      | good-B  | 4        | 200.00     | Beta   |
    When I calculate the cart discount
    Then the discount amount should be 50.00
    And the discount is applied as "4th item 25% off" (1 unit at 200.00, 25% = 50.00)

  Scenario: Buy 3, get 4th with 25% off — only 3 items (no 4th item discount)
    Given a cart with items:
      | good_id | quantity | unit_price | brand  |
      | good-B  | 3        | 200.00     | Beta   |
    When I calculate the cart discount
    Then the discount amount should be 0.00

  Scenario: Buy 3, get 4th with 25% off — 8 items (2 units at 25% off)
    Given a cart with items:
      | good_id | quantity | unit_price | brand  |
      | good-B  | 8        | 200.00     | Beta   |
    When I calculate the cart discount
    Then the discount amount should be 100.00
    And the discount is applied as "4th item 25% off" (2 units: 50.00 + 50.00)
