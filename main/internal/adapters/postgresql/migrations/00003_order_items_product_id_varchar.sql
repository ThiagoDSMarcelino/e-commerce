-- +goose Up
ALTER TABLE order_items
    ALTER COLUMN product_id TYPE VARCHAR(50) USING product_id::text;

-- +goose Down
ALTER TABLE order_items
    ALTER COLUMN product_id TYPE UUID USING product_id::uuid;
