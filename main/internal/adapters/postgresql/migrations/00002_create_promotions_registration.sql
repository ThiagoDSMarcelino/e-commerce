-- +goose Up
CREATE TABLE IF NOT EXISTS promotion_registrations (
    id UUID PRIMARY KEY,
    client_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    email VARCHAR(255) NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS promotion_registrations;
