package core

import "github.com/gofiber/fiber/v2"

// Handler provides shared helpers all domain handlers embed.
type Handler struct{}

func (h *Handler) UserID(c *fiber.Ctx) int {
	id, _ := c.Locals("user_id").(int)
	return id
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
