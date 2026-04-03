CREATE TABLE IF NOT EXISTS raw_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    source VARCHAR(20) NOT NULL,
    raw_text TEXT,
    amount DECIMAL(15,2),
    transaction_type VARCHAR(10),
    detected_at TIMESTAMP,
    metadata JSONB,
    processed BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_raw_user_processed ON raw_transactions(user_id, processed);
CREATE INDEX IF NOT EXISTS idx_raw_created ON raw_transactions(created_at DESC);

CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    raw_transaction_id UUID REFERENCES raw_transactions(id),
    amount DECIMAL(15,2) NOT NULL,
    transaction_type VARCHAR(10) NOT NULL,
    category VARCHAR(50),
    merchant_id UUID,
    description TEXT,
    transaction_date TIMESTAMP NOT NULL,
    balance_after DECIMAL(15,2),
    fingerprint VARCHAR(64) UNIQUE,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_tx_user_date ON transactions(user_id, transaction_date DESC);
CREATE INDEX IF NOT EXISTS idx_tx_category ON transactions(user_id, category);
CREATE INDEX IF NOT EXISTS idx_tx_fingerprint ON transactions(fingerprint);

CREATE TABLE IF NOT EXISTS merchants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    normalized_name VARCHAR(255) NOT NULL,
    category VARCHAR(50),
    aliases TEXT[],
    logo_url VARCHAR(500),
    created_at TIMESTAMP DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_merchant_normalized ON merchants(normalized_name);
