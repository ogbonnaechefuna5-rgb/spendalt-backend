package budget

import (
	"github.com/gofiber/fiber/v2"
	"github.com/moninte/backend/internal/core"
)

type Handler struct {
	core.Handler
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateBudget(c *fiber.Ctx) error {
	var req CreateBudgetRequest
	if err := c.BodyParser(&req); err != nil {
		return h.Fail(c, 400, err)
	}
	b, err := h.service.Create(h.UserID(c), req.Category, req.Amount, req.Period)
	if err != nil {
		return h.Fail(c, 400, err)
	}
	return c.Status(fiber.StatusCreated).JSON(BudgetResponse{Budget: b})
}

func (h *Handler) GetBudgets(c *fiber.Ctx) error {
	page, limit := h.ParsePage(c)
	budgets, err := h.service.GetByUserID(h.UserID(c), limit, (page-1)*limit)
	if err != nil {
		return h.Fail(c, 500, err)
	}
	return c.JSON(BudgetListResponse{Budgets: budgets, PageMeta: core.PageMeta{Page: page, Limit: limit}})
}

func (h *Handler) UpdateBudget(c *fiber.Ctx) error {
	var req UpdateBudgetRequest
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
