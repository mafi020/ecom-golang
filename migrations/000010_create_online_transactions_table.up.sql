CREATE TYPE payment_provider AS ENUM ('stripe', 'paypal', 'sslcommerz');
CREATE TABLE online_transactions (
    id             BIGSERIAL PRIMARY KEY,
    payment_id     BIGINT NOT NULL UNIQUE REFERENCES payments(id) ON DELETE CASCADE,
    provider       payment_provider NOT NULL,
    gateway_ref    VARCHAR(100),                 -- gateway's ID
    gateway_status VARCHAR(50),
    raw_response   JSONB,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_online_transactions_payment_id ON online_transactions(payment_id);