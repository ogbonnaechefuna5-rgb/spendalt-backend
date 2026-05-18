package dashboard

import (
	"github.com/gofiber/fiber/v2"
	"github.com/spendalt/backend/internal/core"
)

type Handler struct {
	core.Handler
	service Service
}

func NewHandler(service Service) *Handler { return &Handler{service: service} }

func (h *Handler) GetDashboard(c *fiber.Ctx) error {
	data, err := h.service.GetDashboard(h.UserID(c))
	if err != nil {
		return h.Fail(c, 500, err)
	}
	return c.JSON(data)
}
