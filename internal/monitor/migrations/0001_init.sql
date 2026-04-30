-- +goose Up
CREATE TABLE IF NOT EXISTS request_logs (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID REFERENCES users(id) ON DELETE SET NULL,
    -- Request
    method         VARCHAR(10)  NOT NULL,
    path           TEXT         NOT NULL,
    status         INT          NOT NULL,
    latency_ms     BIGINT       NOT NULL,
    error          TEXT,
    -- Device
    device_id      VARCHAR(255),
    device_type    VARCHAR(20),
    os             VARCHAR(30),
    app_version    VARCHAR(50),
    user_agent     TEXT,
    -- Network
    ip             VARCHAR(50),
    forwarded_for  TEXT,
    real_ip        VARCHAR(50),
    host           VARCHAR(255),
    protocol       VARCHAR(20),
    tls            BOOLEAN      NOT NULL DEFAULT false,
    origin         TEXT,
    referer        TEXT,
    accept_language VARCHAR(100),
    created_at     TIMESTAMP    NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_request_logs_user_id   ON request_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_request_logs_created_at ON request_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_request_logs_ip         ON request_logs(ip);

-- +goose Down
DROP TABLE IF EXISTS request_logs;
