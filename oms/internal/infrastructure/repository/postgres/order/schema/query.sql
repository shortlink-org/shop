-- name: GetOrder :one
SELECT id, customer_id, status, version, created_at, updated_at
FROM oms.orders
WHERE id = $1;

-- name: GetOrderItems :many
SELECT good_id, quantity, price
FROM oms.order_items
WHERE order_id = $1;

-- name: ListOrdersByCustomer :many
SELECT id, customer_id, status, version, created_at, updated_at
FROM oms.orders
WHERE customer_id = $1
ORDER BY created_at DESC;

-- name: InsertOrder :exec
INSERT INTO oms.orders (id, customer_id, status, version, created_at, updated_at)
VALUES ($1, $2, $3, 1, NOW(), NOW());

-- name: UpdateOrder :execresult
UPDATE oms.orders
SET status = $2, version = $3, updated_at = NOW()
WHERE id = $1 AND version = $4;

-- name: DeleteOrderItems :exec
DELETE FROM oms.order_items
WHERE order_id = $1;

-- name: InsertOrderItem :exec
INSERT INTO oms.order_items (order_id, good_id, quantity, price)
VALUES ($1, $2, $3, $4);
