package budget

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

func (h *Handler) CreateBudget(c *fiber.Ctx) error {
	var req struct {
		Category string  `json:"category"`
		Amount   float64 `json:"amount"`
		Period   string  `json:"period"`
	}
	if err := c.BodyParser(&req); err != nil {
		return h.Fail(c, 400, err)
	}
	b, err := h.service.Create(h.UserID(c), req.Category, req.Amount, req.Period)
	if err != nil {
		return h.Fail(c, 400, err)
	}
	return h.Created(c, "budget", b)
}

func (h *Handler) GetBudgets(c *fiber.Ctx) error {
	page, limit := h.ParsePage(c)
	budgets, err := h.service.GetByUserID(h.UserID(c), limit, (page-1)*limit)
	if err != nil {
		return h.Fail(c, 500, err)
	}
	return c.JSON(fiber.Map{"budgets": budgets, "page": page, "limit": limit})
}

func (h *Handler) UpdateBudget(c *fiber.Ctx) error {
	var req struct {
		Category string  `json:"category"`
		Amount   float64 `json:"amount"`
		Period   string  `json:"period"`
	}
	if err := c.BodyParser(&req); err != nil {
		return h.Fail(c, 400, err)
	}
	if err := h.service.Update(c.Params("id"), h.UserID(c), req.Category, req.Amount, req.Period); err != nil {
		return h.Fail(c, 400, err)
	}
	return h.Message(c, "Budget updated successfully")
}

func (h *Handler) DeleteBudget(c *fiber.Ctx) error {
	if err := h.service.Delete(c.Params("id"), h.UserID(c)); err != nil {
		return h.Fail(c, 400, err)
	}
	return h.Message(c, "Budget deleted successfully")
}
