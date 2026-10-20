-- +goose Up
ALTER TABLE promotion_registrations
    ADD COLUMN categories TEXT[] NOT NULL DEFAULT '{}';

CREATE UNIQUE INDEX IF NOT EXISTS ux_promotion_registrations_client_email
    ON promotion_registrations (client_id, email);

-- +goose Down
DROP INDEX IF EXISTS ux_promotion_registrations_client_email;

ALTER TABLE promotion_registrations DROP COLUMN IF EXISTS categories;
