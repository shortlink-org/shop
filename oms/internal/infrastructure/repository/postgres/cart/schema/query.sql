-- name: GetCart :one
SELECT customer_id, version, created_at, updated_at
FROM oms.carts
WHERE customer_id = $1;

-- name: GetCartItems :many
SELECT good_id, quantity, price, discount
FROM oms.cart_items
WHERE cart_id = $1;

-- name: UpsertCart :execresult
INSERT INTO oms.carts (customer_id, version, created_at, updated_at)
VALUES ($1, $2, NOW(), NOW())
ON CONFLICT (customer_id)
DO UPDATE SET version = $2, updated_at = NOW()
WHERE oms.carts.version = $3;

-- name: InsertCart :exec
INSERT INTO oms.carts (customer_id, version, created_at, updated_at)
VALUES ($1, 1, NOW(), NOW());

-- name: DeleteCartItems :exec
DELETE FROM oms.cart_items
WHERE cart_id = $1;

-- name: InsertCartItem :exec
INSERT INTO oms.cart_items (cart_id, good_id, quantity, price, discount)
VALUES ($1, $2, $3, $4, $5);
