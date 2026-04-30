package monitor

import "time"

// RequestLog is the persisted record of a single inbound request.
type RequestLog struct {
	ID        string    `json:"id"`
	UserID    *string   `json:"user_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`

	// Request / Response
	Method    string  `json:"method"`
	Path      string  `json:"path"`
	Status    int     `json:"status"`
	LatencyMs int64   `json:"latency_ms"`
	Error     *string `json:"error,omitempty"`

	// Device
	DeviceID   string `json:"device_id"`
	DeviceType string `json:"device_type"`
	OS         string `json:"os"`
	AppVersion string `json:"app_version"`
	UserAgent  string `json:"user_agent"`

	// Network
	IP             string `json:"ip"`
	ForwardedFor   string `json:"forwarded_for"`
	RealIP         string `json:"real_ip"`
	Host           string `json:"host"`
	Protocol       string `json:"protocol"`
	TLS            bool   `json:"tls"`
	Origin         string `json:"origin"`
	Referer        string `json:"referer"`
	AcceptLanguage string `json:"accept_language"`
}
