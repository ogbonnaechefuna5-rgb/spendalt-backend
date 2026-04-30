package monitor

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/spendalt/backend/internal/core"
)

type Handler struct {
	core.Handler
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

// GetActivity returns the most recent request logs for the authenticated user.
// Query param: limit (default 20, max 100).
func (h *Handler) GetActivity(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	logs, err := h.repo.GetByUserID(h.UserID(c), limit)
	if err != nil {
		return h.Fail(c, 500, err)
	}
	return h.OK(c, "activity", logs)
}
