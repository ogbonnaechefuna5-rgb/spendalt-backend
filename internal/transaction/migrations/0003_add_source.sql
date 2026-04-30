-- +goose Up
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS source VARCHAR(20) NOT NULL DEFAULT 'manual';

-- +goose Down
ALTER TABLE transactions DROP COLUMN IF EXISTS source;
