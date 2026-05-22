-- +goose Up
CREATE TABLE IF NOT EXISTS http_logs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id      VARCHAR(36)  NOT NULL DEFAULT '',
    user_id         UUID REFERENCES users(id) ON DELETE SET NULL,
    -- Request / Response
    method          VARCHAR(10)  NOT NULL,
    path            TEXT         NOT NULL,
    status          INT          NOT NULL,
    latency_ms      BIGINT       NOT NULL,
    response_body   TEXT,
    error           TEXT,
    -- Device
    device_id       VARCHAR(255),
    device_type     VARCHAR(20),
    os              VARCHAR(30),
    app_version     VARCHAR(50),
    user_agent      TEXT,
    -- Network
    ip              VARCHAR(50),
    forwarded_for   TEXT,
    real_ip         VARCHAR(50),
    host            VARCHAR(255),
    protocol        VARCHAR(20),
    tls             BOOLEAN      NOT NULL DEFAULT false,
    origin          TEXT,
    referer         TEXT,
    accept_language VARCHAR(100),
    created_at      TIMESTAMP    NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_http_logs_user_id    ON http_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_http_logs_created_at ON http_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_http_logs_ip         ON http_logs(ip);
CREATE INDEX IF NOT EXISTS idx_http_logs_request_id ON http_logs(request_id);

-- +goose Down
DROP TABLE IF EXISTS http_logs;
