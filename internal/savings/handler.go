package savings

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/spendalt/backend/internal/core"
)

type Handler struct {
	core.Handler
	service Service
}

func NewHandler(service Service) *Handler { return &Handler{service: service} }

func (h *Handler) CreateGoal(c *fiber.Ctx) error {
	var req struct {
		Name         string     `json:"name"`
		TargetAmount float64    `json:"target_amount"`
		Deadline     *time.Time `json:"deadline"`
	}
	if err := c.BodyParser(&req); err != nil {
		return h.Fail(c, 400, err)
	}
	g, err := h.service.Create(h.UserID(c), req.Name, req.TargetAmount, req.Deadline)
	if err != nil {
		return h.Fail(c, 400, err)
	}
	return h.Created(c, "goal", g)
}

func (h *Handler) GetGoals(c *fiber.Ctx) error {
	page, limit := h.ParsePage(c)
	goals, err := h.service.GetByUserID(h.UserID(c), limit, (page-1)*limit)
	if err != nil {
		return h.Fail(c, 500, err)
	}
	return c.JSON(fiber.Map{"goals": goals, "page": page, "limit": limit})
}

func (h *Handler) UpdateProgress(c *fiber.Ctx) error {
	var req struct {
		Amount float64 `json:"amount"`
	}
	if err := c.BodyParser(&req); err != nil {
		return h.Fail(c, 400, err)
	}
	if err := h.service.UpdateProgress(c.Params("id"), h.UserID(c), req.Amount); err != nil {
		return h.Fail(c, 400, err)
	}
	return h.Message(c, "Progress updated")
}

func (h *Handler) DeleteGoal(c *fiber.Ctx) error {
	if err := h.service.Delete(c.Params("id"), h.UserID(c)); err != nil {
		return h.Fail(c, 400, err)
	}
	return h.Message(c, "Goal deleted")
}
