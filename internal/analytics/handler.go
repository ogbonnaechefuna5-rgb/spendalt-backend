package analytics

import (
	"github.com/gofiber/fiber/v2"
	"github.com/spendalt/backend/internal/core"
)

type Handler struct {
	core.Handler
	service Service
}

func NewHandler(service Service) *Handler { return &Handler{service: service} }

func (h *Handler) GetInsights(c *fiber.Ctx) error {
	ins, err := h.service.GetInsights(h.UserID(c))
	if err != nil {
		return h.Fail(c, 500, err)
	}
	return h.OK(c, "insights", ins)
}

func (h *Handler) GetWeeklyTrend(c *fiber.Ctx) error {
	trend, err := h.service.GetWeeklyTrend(h.UserID(c))
	if err != nil {
		return h.Fail(c, 500, err)
	}
	return h.OK(c, "trend", trend)
}

func (h *Handler) GetHealthScore(c *fiber.Ctx) error {
	score, err := h.service.GetHealthScore(h.UserID(c))
	if err != nil {
		return h.Fail(c, 500, err)
	}
	return h.OK(c, "health", score)
}
