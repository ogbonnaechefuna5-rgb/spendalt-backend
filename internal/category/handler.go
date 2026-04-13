package category

import (
	"github.com/gofiber/fiber/v2"
	"github.com/spendalt/backend/internal/core"
)

type Handler struct {
	core.Handler
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetCategories(c *fiber.Ctx) error {
	page, limit := h.ParsePage(c)
	categories, err := h.service.GetCategories(limit, (page-1)*limit)
	if err != nil {
		return h.Fail(c, 500, err)
	}
	return c.JSON(fiber.Map{"categories": categories, "page": page, "limit": limit})
}

func (h *Handler) GetCategoryBreakdown(c *fiber.Ctx) error {
	page, limit := h.ParsePage(c)
	breakdown, err := h.service.GetCategoryBreakdown(h.UserID(c), limit, (page-1)*limit)
	if err != nil {
		return h.Fail(c, 500, err)
	}
	return c.JSON(fiber.Map{"breakdown": breakdown, "page": page, "limit": limit})
}
