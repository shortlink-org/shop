-- OMS Order Schema
CREATE SCHEMA IF NOT EXISTS oms;

-- Orders table with optimistic locking via version field
CREATE TABLE IF NOT EXISTS oms.orders (
    id          UUID PRIMARY KEY,
    customer_id UUID NOT NULL,
    status      VARCHAR(32) NOT NULL DEFAULT 'PENDING',
    version     INT NOT NULL DEFAULT 1,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE oms.orders IS 'Customer orders';
COMMENT ON COLUMN oms.orders.version IS 'Optimistic concurrency control version';
COMMENT ON COLUMN oms.orders.status IS 'Order status: PENDING, PROCESSING, COMPLETED, CANCELLED';

CREATE INDEX IF NOT EXISTS orders_customer_id_idx ON oms.orders(customer_id);
CREATE INDEX IF NOT EXISTS orders_status_idx ON oms.orders(status);

-- Order items
CREATE TABLE IF NOT EXISTS oms.order_items (
    order_id  UUID NOT NULL REFERENCES oms.orders(id) ON DELETE CASCADE,
    good_id   UUID NOT NULL,
    quantity  INT NOT NULL CHECK (quantity > 0),
    price     DECIMAL(12,2) NOT NULL,
    PRIMARY KEY (order_id, good_id)
);

COMMENT ON TABLE oms.order_items IS 'Items in orders';

CREATE INDEX IF NOT EXISTS order_items_order_id_idx ON oms.order_items(order_id);
