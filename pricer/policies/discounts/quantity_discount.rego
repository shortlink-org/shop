package pricing.discount

# "3 for 2" style discount: when quantity >= min_quantity, one unit free per set
# e.g. min_quantity=3: buy 3 get 1 free
quantity_discount[item.productId] = discount {
	item := input.items[_]
	min_qty := input.params.min_quantity_for_discount
	item.quantity >= min_qty
	sets := floor(item.quantity / min_qty)
	discount := sets * item.price
}

total_quantity_discount := sum([discount | discount := quantity_discount[_]])
