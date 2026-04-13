package core

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// Handler provides shared helpers all domain handlers embed.
type Handler struct{}

func (h *Handler) UserID(c *fiber.Ctx) string {
	id, _ := c.Locals("user_id").(string)
	return id
}

// ParsePage parses and clamps page/limit query params.
// Defaults: page=1, limit=20. Max limit=100.
func (h *Handler) ParsePage(c *fiber.Ctx) (page, limit int) {
	page, _ = strconv.Atoi(c.Query("page", "1"))
	limit, _ = strconv.Atoi(c.Query("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return
}

func (h *Handler) OK(c *fiber.Ctx, key string, data interface{}) error {
	return c.JSON(fiber.Map{key: data})
}

func (h *Handler) Created(c *fiber.Ctx, key string, data interface{}) error {
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{key: data})
}

func (h *Handler) Fail(c *fiber.Ctx, status int, err error) error {
	return c.Status(status).JSON(fiber.Map{"error": err.Error()})
}

func (h *Handler) Message(c *fiber.Ctx, msg string) error {
	return c.JSON(fiber.Map{"message": msg})
}
