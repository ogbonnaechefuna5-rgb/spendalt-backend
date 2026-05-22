-- +goose Up
ALTER TABLE transactions DROP COLUMN IF EXISTS merchant_id;
DROP TABLE IF EXISTS merchants;

-- +goose Down
CREATE TABLE IF NOT EXISTS merchants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    normalized_name VARCHAR(255) NOT NULL,
    category VARCHAR(50),
    aliases TEXT[],
    logo_url VARCHAR(500),
    created_at TIMESTAMP DEFAULT NOW()
);
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS merchant_id UUID;
