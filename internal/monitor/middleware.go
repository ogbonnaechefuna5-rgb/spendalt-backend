package monitor

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

// RequestMonitor logs and persists every inbound request/response with full device and network info.
func RequestMonitor(log Logger, repo Repository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()

		d := ExtractDeviceInfo(c)
		latency := time.Since(start).Milliseconds()
		status := c.Response().StatusCode()

		entry := &RequestLog{
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

		// Attach authenticated user if present
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
