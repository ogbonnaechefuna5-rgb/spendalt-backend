-- +goose Up
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS merchant_name VARCHAR(255);

-- +goose Down
-- intentionally left empty; dropping a column is destructive
