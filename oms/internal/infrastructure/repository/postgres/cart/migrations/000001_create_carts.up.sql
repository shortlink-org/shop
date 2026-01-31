-- OMS Cart Schema
CREATE SCHEMA IF NOT EXISTS oms;

-- Carts table with optimistic locking via version field
CREATE TABLE IF NOT EXISTS oms.carts (
    customer_id UUID PRIMARY KEY,
    version     INT NOT NULL DEFAULT 1,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE oms.carts IS 'Shopping carts by customer';
COMMENT ON COLUMN oms.carts.version IS 'Optimistic concurrency control version';

-- Cart items
CREATE TABLE IF NOT EXISTS oms.cart_items (
    cart_id   UUID NOT NULL REFERENCES oms.carts(customer_id) ON DELETE CASCADE,
    good_id   UUID NOT NULL,
    quantity  INT NOT NULL CHECK (quantity > 0),
    price     DECIMAL(12,2) NOT NULL,
    discount  DECIMAL(12,2) NOT NULL DEFAULT 0 CHECK (discount >= 0),
    PRIMARY KEY (cart_id, good_id)
);

COMMENT ON TABLE oms.cart_items IS 'Items in shopping carts';

CREATE INDEX IF NOT EXISTS cart_items_cart_id_idx ON oms.cart_items(cart_id);
