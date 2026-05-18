package monitor

import "github.com/gofiber/fiber/v2"

// SecurityHeaders sets defensive HTTP headers on every response.
func SecurityHeaders() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("X-XSS-Protection", "0") // modern browsers: disable legacy XSS auditor (CSP is the right tool)
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		c.Set("Content-Security-Policy", "default-src 'none'")
		c.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		return c.Next()
	}
}
