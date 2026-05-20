package monitor

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/moninte/backend/internal/common"
)

const RequestIDHeader = "X-Request-ID"
const RequestIDLocal = "request_id"

// RequestMonitor generates a unique request_id for every inbound request,
// sets it on the response header, logs the full request/response, and
// persists the entry asynchronously.
func RequestMonitor(log Logger, repo Repository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Accept a client-supplied ID (e.g. from mobile app) or generate one.
		requestID := c.Get(RequestIDHeader)
		if requestID == "" {
			requestID = common.NewID()
		}
		c.Locals(RequestIDLocal, requestID)
		c.Set(RequestIDHeader, requestID)

		start := time.Now()
		err := c.Next()

		d := ExtractDeviceInfo(c)
		latency := time.Since(start).Milliseconds()
		status := c.Response().StatusCode()

		entry := &RequestLog{
			RequestID: requestID,
			Method:    c.Method(),
			Path:      c.Path(),
			Status:    status,
			LatencyMs: latency,
			// Device
			DeviceID:   d.DeviceID,
			DeviceType: d.DeviceType,
			OS:         d.OS,
			AppVersion: d.AppVersion,
			UserAgent:  d.UserAgent,
			// Network
			IP:             d.IP,
			ForwardedFor:   d.ForwardedFor,
			RealIP:         d.RealIP,
			Host:           d.Host,
			Protocol:       d.Protocol,
			TLS:            d.TLS,
			Origin:         d.Origin,
			Referer:        d.Referer,
			AcceptLanguage: d.AcceptLanguage,
		}

		if uid, ok := c.Locals("user_id").(string); ok && uid != "" {
			entry.UserID = &uid
		}

		if err != nil {
			msg := err.Error()
			entry.Error = &msg
			log.Error("request", requestArgs(entry)...)
		} else {
			log.Info("request", requestArgs(entry)...)
		}

		repo.Save(entry)
		return err
	}
}

func requestArgs(e *RequestLog) []any {
	return []any{
		"request_id", e.RequestID,
		"method", e.Method,
		"path", e.Path,
		"status", e.Status,
		"latency_ms", e.LatencyMs,
		"device_id", e.DeviceID,
		"device_type", e.DeviceType,
		"os", e.OS,
		"app_version", e.AppVersion,
		"user_agent", e.UserAgent,
		"ip", e.IP,
		"forwarded_for", e.ForwardedFor,
		"real_ip", e.RealIP,
		"host", e.Host,
		"protocol", e.Protocol,
		"tls", e.TLS,
		"origin", e.Origin,
		"referer", e.Referer,
		"accept_language", e.AcceptLanguage,
	}
}
