-- +goose Up
CREATE TABLE IF NOT EXISTS products (
    id     VARCHAR(50)  PRIMARY KEY,
    name   VARCHAR(255) NOT NULL,
    amount INTEGER      NOT NULL CHECK (amount >= 0)
);

-- +goose Down
DROP TABLE IF EXISTS products;
