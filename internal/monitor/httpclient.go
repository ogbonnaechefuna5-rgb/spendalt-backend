package monitor

import (
	"fmt"
	"net/http"
	"time"
)

// monitoredTransport wraps any http.RoundTripper to log outbound calls (OCP).
type monitoredTransport struct {
	next http.RoundTripper
	log  Logger
}

// NewHTTPClient returns an *http.Client whose transport logs every outbound
// request and response. Pass nil to use http.DefaultTransport.
func NewHTTPClient(next http.RoundTripper, log Logger) *http.Client {
	if next == nil {
		next = http.DefaultTransport
	}
	return &http.Client{Transport: &monitoredTransport{next: next, log: log}}
}

func (t *monitoredTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
	resp, err := t.next.RoundTrip(req)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		t.log.Error("outbound request failed",
			"method", req.Method,
			"url", req.URL.String(),
			"latency_ms", latency,
			"error", err.Error(),
		)
		return nil, err
	}

	t.log.Info("outbound request",
		"method", req.Method,
		"url", req.URL.String(),
		"status", fmt.Sprintf("%d %s", resp.StatusCode, http.StatusText(resp.StatusCode)),
		"latency_ms", latency,
	)
	return resp, nil
}
