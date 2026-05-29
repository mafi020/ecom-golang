CREATE TYPE payment_method AS ENUM ('cod', 'online');
CREATE TYPE payment_status AS ENUM ('pending', 'completed', 'failed', 'refunded');

CREATE TABLE payments (
    id         BIGSERIAL PRIMARY KEY,
    order_id   BIGINT NOT NULL UNIQUE REFERENCES orders(id) ON DELETE RESTRICT,
    transaction_id VARCHAR(100) NOT NULL UNIQUE, -- UUID for idempotency and internal tracking
    method     payment_method NOT NULL,
    status     payment_status NOT NULL DEFAULT 'pending',
    amount     NUMERIC(10, 2) NOT NULL CHECK (amount > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payments_order_id              ON payments(order_id);
CREATE INDEX idx_payments_transaction_id        ON payments(transaction_id);
CREATE INDEX idx_payments_status                ON payments(status);
CREATE INDEX idx_payments_method                ON payments(method);





