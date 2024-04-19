CREATE TABLE models (
    id BIGSERIAL PRIMARY KEY,
    order_uid VARCHAR UNIQUE NOT NULL,
    model JSONB
);

CREATE INDEX idx_order_uid ON models USING HASH (order_uid);