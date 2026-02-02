package pricing.discount

# Combination discount: apply when cart has 2+ different products
# percent_discount is applied to cart subtotal (sum of item.price * quantity)
default total_combination_discount = 0

total_combination_discount = discount {
	count(input.items) >= 2
	subtotal := sum([prod | item := input.items[_]; prod := item.price * item.quantity])
	percent := input.params.combination_discount_percent
	discount := subtotal * percent
}
