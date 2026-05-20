-- +goose Up
ALTER TABLE request_logs ADD COLUMN IF NOT EXISTS request_id VARCHAR(36) NOT NULL DEFAULT '';
CREATE INDEX IF NOT EXISTS idx_request_logs_request_id ON request_logs(request_id);

-- +goose Down
DROP INDEX IF EXISTS idx_request_logs_request_id;
ALTER TABLE request_logs DROP COLUMN IF EXISTS request_id;
