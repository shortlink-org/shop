package pricing.discount

# Test: 5% combination discount when 2+ products (600+700)*0.05 = 65
test_combination_discount_applied if {
	total_combination_discount == 65.0 with input as {
		"items": [
			{"productId": "a", "quantity": 1, "price": 600},
			{"productId": "b", "quantity": 1, "price": 700}
		],
		"params": {"combination_discount_percent": 0.05}
	}
}

# Test: no combination discount for single product
test_combination_discount_single_product if {
	total_combination_discount == 0 with input as {
		"items": [{"productId": "a", "quantity": 2, "price": 100}],
		"params": {"combination_discount_percent": 0.05}
	}
}
