package monitor

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// DeviceInfo holds everything extractable about the client device and network.
type DeviceInfo struct {
	// Device
	UserAgent    string
	DeviceID     string // X-Device header
	DeviceType   string // mobile | tablet | desktop | unknown
	OS           string
	AppVersion   string // X-App-Version header

	// Network
	IP             string
	ForwardedFor   string
	RealIP         string
	Protocol       string // HTTP/1.1, HTTP/2, etc.
	TLS            bool
	Referer        string
	Origin         string
	AcceptLanguage string
	Host           string
}

func ExtractDeviceInfo(c *fiber.Ctx) DeviceInfo {
	ua := c.Get(fiber.HeaderUserAgent)
	return DeviceInfo{
		UserAgent:      ua,
		DeviceID:       c.Get("X-Device"),
		DeviceType:     inferDeviceType(ua),
		OS:             inferOS(ua),
		AppVersion:     c.Get("X-App-Version"),
		IP:             c.IP(),
		ForwardedFor:   c.Get(fiber.HeaderXForwardedFor),
		RealIP:         c.Get("X-Real-IP"),
		Protocol:       c.Protocol(),
		TLS:            c.Secure(),
		Referer:        c.Get(fiber.HeaderReferer),
		Origin:         c.Get(fiber.HeaderOrigin),
		AcceptLanguage: c.Get(fiber.HeaderAcceptLanguage),
		Host:           c.Hostname(),
	}
}

func inferDeviceType(ua string) string {
	ua = strings.ToLower(ua)
	switch {
	case strings.Contains(ua, "tablet") || strings.Contains(ua, "ipad"):
		return "tablet"
	case strings.Contains(ua, "mobile") || strings.Contains(ua, "android") ||
		strings.Contains(ua, "iphone") || strings.Contains(ua, "okhttp") ||
		strings.Contains(ua, "cfnetwork"):
		return "mobile"
	case ua == "":
		return "unknown"
	default:
		return "desktop"
	}
}

func inferOS(ua string) string {
	ua = strings.ToLower(ua)
	switch {
	case strings.Contains(ua, "android") || strings.Contains(ua, "okhttp"):
		return "android"
	case strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad") ||
		strings.Contains(ua, "ios") || strings.Contains(ua, "cfnetwork") ||
		strings.Contains(ua, "darwin"):
		return "ios"
	case strings.Contains(ua, "windows"):
		return "windows"
	case strings.Contains(ua, "mac"):
		return "macos"
	case strings.Contains(ua, "linux"):
		return "linux"
	default:
		return "unknown"
	}
}
