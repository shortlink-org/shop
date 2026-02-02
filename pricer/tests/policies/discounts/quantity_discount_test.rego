package pricing.discount

# Test: 3-for-2 discount - buy 3 get 1 free
test_quantity_discount_three_for_two if {
	quantity_discount["item1"] == 59.99 with input as {
		"items": [{"productId": "item1", "quantity": 3, "price": 59.99}],
		"params": {"min_quantity_for_discount": 3}
	}
}

# Test: no discount when quantity < min
test_quantity_discount_below_min if {
	count(quantity_discount) == 0 with input as {
		"items": [{"productId": "item1", "quantity": 2, "price": 100}],
		"params": {"min_quantity_for_discount": 3}
	}
}
