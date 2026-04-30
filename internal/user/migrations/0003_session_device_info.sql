-- +goose Up
ALTER TABLE user_sessions
    ADD COLUMN IF NOT EXISTS device_type  VARCHAR(20),
    ADD COLUMN IF NOT EXISTS os           VARCHAR(30),
    ADD COLUMN IF NOT EXISTS app_version  VARCHAR(50);

-- +goose Down
ALTER TABLE user_sessions
    DROP COLUMN IF EXISTS device_type,
    DROP COLUMN IF EXISTS os,
    DROP COLUMN IF EXISTS app_version;
