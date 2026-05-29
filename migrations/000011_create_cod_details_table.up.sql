CREATE TYPE cod_delivery_status AS ENUM ('pending', 'shipped', 'delivered', 'collected');

CREATE TABLE cod_details (
    id              BIGSERIAL PRIMARY KEY,
    payment_id      BIGINT NOT NULL UNIQUE REFERENCES payments(id) ON DELETE CASCADE,
    collected_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_cod_details_payment_id         ON cod_details(payment_id);