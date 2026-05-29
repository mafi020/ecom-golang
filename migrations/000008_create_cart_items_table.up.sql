CREATE TABLE cart_items (
    id         BIGSERIAL PRIMARY KEY,
    cart_id    BIGINT NOT NULL REFERENCES carts(id) ON DELETE CASCADE,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    quantity   INT NOT NULL CHECK (quantity > 0),
    price      NUMERIC(10, 2) NOT NULL CHECK (price >= 0),
    UNIQUE (cart_id, product_id)
);

CREATE INDEX idx_cart_items_cart_id    ON cart_items(cart_id);
CREATE INDEX idx_cart_items_product_id ON cart_items(product_id);