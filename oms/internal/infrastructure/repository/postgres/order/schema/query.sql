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

-- name: ListOrders :many
SELECT id, customer_id, status, version, created_at, updated_at
FROM oms.orders
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListOrdersWithCustomerFilter :many
SELECT id, customer_id, status, version, created_at, updated_at
FROM oms.orders
WHERE customer_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListOrdersWithStatusFilter :many
SELECT id, customer_id, status, version, created_at, updated_at
FROM oms.orders
WHERE status = ANY($1::int[])
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListOrdersWithFilters :many
SELECT id, customer_id, status, version, created_at, updated_at
FROM oms.orders
WHERE customer_id = $1 AND status = ANY($2::int[])
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountOrders :one
SELECT COUNT(*) FROM oms.orders;

-- name: CountOrdersByCustomer :one
SELECT COUNT(*) FROM oms.orders WHERE customer_id = $1;

-- name: CountOrdersByStatus :one
SELECT COUNT(*) FROM oms.orders WHERE status = ANY($1::int[]);

-- name: CountOrdersWithFilters :one
SELECT COUNT(*) FROM oms.orders WHERE customer_id = $1 AND status = ANY($2::int[]);

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

-- name: GetOrderDeliveryInfo :one
SELECT 
    order_id,
    pickup_street, pickup_city, pickup_postal_code, pickup_country, pickup_latitude, pickup_longitude,
    delivery_street, delivery_city, delivery_postal_code, delivery_country, delivery_latitude, delivery_longitude,
    period_start, period_end,
    weight_kg, dimensions,
    priority, package_id
FROM oms.order_delivery_info
WHERE order_id = $1;

-- name: InsertOrderDeliveryInfo :exec
INSERT INTO oms.order_delivery_info (
    order_id,
    pickup_street, pickup_city, pickup_postal_code, pickup_country, pickup_latitude, pickup_longitude,
    delivery_street, delivery_city, delivery_postal_code, delivery_country, delivery_latitude, delivery_longitude,
    period_start, period_end,
    weight_kg, dimensions,
    priority, package_id
) VALUES (
    $1,
    $2, $3, $4, $5, $6, $7,
    $8, $9, $10, $11, $12, $13,
    $14, $15,
    $16, $17,
    $18, $19
);

-- name: UpdateOrderDeliveryInfo :exec
UPDATE oms.order_delivery_info
SET 
    pickup_street = $2, pickup_city = $3, pickup_postal_code = $4, pickup_country = $5, pickup_latitude = $6, pickup_longitude = $7,
    delivery_street = $8, delivery_city = $9, delivery_postal_code = $10, delivery_country = $11, delivery_latitude = $12, delivery_longitude = $13,
    period_start = $14, period_end = $15,
    weight_kg = $16, dimensions = $17,
    priority = $18, package_id = $19
WHERE order_id = $1;

-- name: DeleteOrderDeliveryInfo :exec
DELETE FROM oms.order_delivery_info
WHERE order_id = $1;
