package notification

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/moninte/backend/internal/core"
	"github.com/moninte/backend/internal/lang"
)

// Handler exposes notification endpoints.
type Handler struct {
	core.Handler
	service Service
}

// NewHandler returns a new Handler.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// List handles GET /notifications
func (h *Handler) List(c *fiber.Ctx) error {
	page, limit := h.ParsePage(c)
	items, err := h.service.List(h.UserID(c), page, limit)
	if err != nil {
		return h.Fail(c, 500, err)
	}
	unread, err := h.service.UnreadCount(h.UserID(c))
	if err != nil {
		return h.Fail(c, 500, err)
	}
	if items == nil {
		items = []*Notification{}
	}
	return c.JSON(ListResponse{
		Notifications: items,
		UnreadCount:   unread,
		PageMeta:      core.PageMeta{Page: page, Limit: limit},
	})
}

// MarkRead handles POST /notifications/:id/read
func (h *Handler) MarkRead(c *fiber.Ctx) error {
	if err := h.service.MarkRead(c.Params("id"), h.UserID(c)); err != nil {
		return h.Fail(c, 400, err)
	}
	return h.Message(c, "notification marked as read")
}

// MarkAllRead handles POST /notifications/read-all
func (h *Handler) MarkAllRead(c *fiber.Ctx) error {
	if err := h.service.MarkAllRead(h.UserID(c)); err != nil {
		return h.Fail(c, 500, err)
	}
	return h.Message(c, "all notifications marked as read")
}

// Delete handles DELETE /notifications/:id
func (h *Handler) Delete(c *fiber.Ctx) error {
	if err := h.service.Delete(c.Params("id"), h.UserID(c)); err != nil {
		return h.Fail(c, 400, err)
	}
	return h.Message(c, "notification deleted")
}

// RegisterToken handles POST /user/device-token
// Body: { "token": "...", "platform": "android"|"ios"|"web" }
func (h *Handler) RegisterToken(c *fiber.Ctx) error {
	var req struct {
		Token    string `json:"token"`
		Platform string `json:"platform"`
	}
	if err := c.BodyParser(&req); err != nil || req.Token == "" {
		return h.Fail(c, 400, errors.New(lang.ErrInvalidBody))
	}
	if req.Platform == "" {
		req.Platform = "android"
	}
	if err := h.service.RegisterToken(h.UserID(c), req.Token, req.Platform); err != nil {
		return h.Fail(c, 500, err)
	}
	return h.Message(c, "device token registered")
}

// RemoveToken handles DELETE /user/device-token
// Body: { "token": "..." }
func (h *Handler) RemoveToken(c *fiber.Ctx) error {
	var req struct {
		Token string `json:"token"`
	}
	if err := c.BodyParser(&req); err != nil || req.Token == "" {
		return h.Fail(c, 400, errors.New(lang.ErrInvalidBody))
	}
	if err := h.service.RemoveToken(h.UserID(c), req.Token); err != nil {
		return h.Fail(c, 500, err)
	}
	return h.Message(c, "device token removed")
}
