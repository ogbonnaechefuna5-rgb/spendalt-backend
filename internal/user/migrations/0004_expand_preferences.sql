-- +goose Up
ALTER TABLE user_preferences
    ADD COLUMN IF NOT EXISTS transaction_alerts  BOOLEAN DEFAULT true,
    ADD COLUMN IF NOT EXISTS budget_warnings     BOOLEAN DEFAULT true,
    ADD COLUMN IF NOT EXISTS ai_insights         BOOLEAN DEFAULT true,
    ADD COLUMN IF NOT EXISTS weekly_report       BOOLEAN DEFAULT false,
    ADD COLUMN IF NOT EXISTS savings_reminders   BOOLEAN DEFAULT true,
    ADD COLUMN IF NOT EXISTS promotions          BOOLEAN DEFAULT false,
    ADD COLUMN IF NOT EXISTS hide_balances       BOOLEAN DEFAULT false,
    ADD COLUMN IF NOT EXISTS crash_reports       BOOLEAN DEFAULT true;

-- +goose Down
ALTER TABLE user_preferences
    DROP COLUMN IF EXISTS transaction_alerts,
    DROP COLUMN IF EXISTS budget_warnings,
    DROP COLUMN IF EXISTS ai_insights,
    DROP COLUMN IF EXISTS weekly_report,
    DROP COLUMN IF EXISTS savings_reminders,
    DROP COLUMN IF EXISTS promotions,
    DROP COLUMN IF EXISTS hide_balances,
    DROP COLUMN IF EXISTS crash_reports;
