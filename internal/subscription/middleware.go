package subscription

import (
	"github.com/gofiber/fiber/v2"
	"github.com/moninte/backend/internal/core"
	"github.com/moninte/backend/internal/lang"
)

// RequireEntitlement gates a route to users whose active plan includes the given feature.
func RequireEntitlement(svc Service, feature string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, _ := c.Locals("user_id").(string)
		ok, err := svc.HasEntitlement(userID, feature)
		if err != nil {
			return c.Status(500).JSON(core.ErrorResponse{Error: lang.ErrInternal})
		}
		if !ok {
			return c.Status(403).JSON(core.ErrorResponse{Error: "this feature requires a premium plan"})
		}
		return c.Next()
	}
}
