CREATE TYPE order_status AS ENUM (
    'pending',
    'paid',
    'shipped',
    'delivered',
    'completed',
    'cancelled'
);

CREATE TABLE orders (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    status          order_status NOT NULL DEFAULT 'pending',
    total_price     NUMERIC(10, 2) NOT NULL CHECK (total_price >= 0),
    
    -- Logistics Data Fields (Moved from Payment/COD tracking domains)
    courier_partner VARCHAR(255),
    tracking_number VARCHAR(255),
    shipped_at      TIMESTAMPTZ,
    delivered_at    TIMESTAMPTZ,
    
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Core Indexes
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status  ON orders(status);

-- Operational Tracking Indexes (Ensures instant warehouse & customer lookups)
CREATE INDEX idx_orders_tracking_number ON orders(tracking_number) WHERE tracking_number IS NOT NULL;
CREATE INDEX idx_orders_courier_partner ON orders(courier_partner) WHERE courier_partner IS NOT NULL;
