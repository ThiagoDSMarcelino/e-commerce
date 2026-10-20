-- +goose Up
ALTER TABLE orders
    ADD COLUMN payment_url TEXT;

-- +goose Down
ALTER TABLE orders
    DROP COLUMN payment_url;
