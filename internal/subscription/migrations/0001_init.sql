-- +goose Up

CREATE TYPE subscription_status AS ENUM ('active', 'cancelled', 'expired', 'past_due');

CREATE TABLE IF NOT EXISTS plans (
    id          VARCHAR(50) PRIMARY KEY,  -- e.g. 'free', 'premium', 'pro', 'business'
    name        VARCHAR(100) NOT NULL,
    description TEXT,
    price_ngn   DECIMAL(15,2) NOT NULL DEFAULT 0,
    interval    VARCHAR(20) NOT NULL DEFAULT 'monthly', -- monthly, yearly, one_time
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS plan_entitlements (
    plan_id     VARCHAR(50) NOT NULL REFERENCES plans(id) ON DELETE CASCADE,
    feature     VARCHAR(100) NOT NULL, -- e.g. 'mono_sync', 'ml_insights', 'pdf_export', 'statement_upload'
    PRIMARY KEY (plan_id, feature)
);

CREATE TABLE IF NOT EXISTS subscriptions (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id                 UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan_id                 VARCHAR(50) NOT NULL REFERENCES plans(id),
    status                  subscription_status NOT NULL DEFAULT 'active',
    provider                VARCHAR(50),          -- 'paystack', 'flutterwave', 'manual', etc.
    provider_reference      VARCHAR(255),         -- provider's subscription/transaction ID
    current_period_start    TIMESTAMP NOT NULL DEFAULT NOW(),
    current_period_end      TIMESTAMP,            -- NULL = lifetime / free
    cancelled_at            TIMESTAMP,
    created_at              TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_status  ON subscriptions(status);

-- Seed default plans
INSERT INTO plans (id, name, description, price_ngn, interval) VALUES
    ('free',     'Free',     'Basic personal finance tracking',                    0,      'monthly'),
    ('premium',  'Premium',  'Mono bank sync + advanced analytics',                2000,   'monthly'),
    ('pro',      'Pro',      'Premium + ML insights + PDF export',                 5000,   'monthly'),
    ('business', 'Business', 'Pro + multi-account + team members',                 15000,  'monthly')
ON CONFLICT (id) DO NOTHING;

-- Seed entitlements
INSERT INTO plan_entitlements (plan_id, feature) VALUES
    ('premium',  'mono_sync'),
    ('premium',  'statement_upload'),
    ('pro',      'mono_sync'),
    ('pro',      'statement_upload'),
    ('pro',      'ml_insights'),
    ('pro',      'pdf_export'),
    ('business', 'mono_sync'),
    ('business', 'statement_upload'),
    ('business', 'ml_insights'),
    ('business', 'pdf_export'),
    ('business', 'multi_account'),
    ('business', 'team_members')
ON CONFLICT DO NOTHING;

-- Every new user gets a free subscription automatically
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION create_free_subscription()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO subscriptions (user_id, plan_id, status, provider)
    VALUES (NEW.id, 'free', 'active', 'system');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER trg_user_free_subscription
    AFTER INSERT ON users
    FOR EACH ROW EXECUTE FUNCTION create_free_subscription();

-- +goose Down
DROP TRIGGER IF EXISTS trg_user_free_subscription ON users;
DROP FUNCTION IF EXISTS create_free_subscription();
DROP TABLE IF EXISTS plan_entitlements;
DROP TABLE IF EXISTS subscriptions;
DROP TABLE IF EXISTS plans;
DROP TYPE IF EXISTS subscription_status;
