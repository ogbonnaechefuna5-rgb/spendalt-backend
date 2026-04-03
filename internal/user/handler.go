package user

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

func (h *Handler) GetProfile(c *fiber.Ctx) error {
	user, err := h.service.GetProfile(h.UserID(c))
	if err != nil {
		return h.Fail(c, 404, err)
	}
	return h.OK(c, "user", user)
}

func (h *Handler) UpdateProfile(c *fiber.Ctx) error {
	var req struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Phone     string `json:"phone"`
	}
	if err := c.BodyParser(&req); err != nil {
		return h.Fail(c, 400, err)
	}
	if err := h.service.UpdateProfile(h.UserID(c), req.FirstName, req.LastName, req.Phone); err != nil {
		return h.Fail(c, 400, err)
	}
	return h.Message(c, "Profile updated successfully")
}

func (h *Handler) ChangePassword(c *fiber.Ctx) error {
	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := c.BodyParser(&req); err != nil {
		return h.Fail(c, 400, err)
	}
	if err := h.service.ChangePassword(h.UserID(c), req.OldPassword, req.NewPassword); err != nil {
		return h.Fail(c, 400, err)
	}
	return h.Message(c, "Password changed successfully")
}

func (h *Handler) DeleteAccount(c *fiber.Ctx) error {
	if err := h.service.DeleteAccount(h.UserID(c)); err != nil {
		return h.Fail(c, 400, err)
	}
	return h.Message(c, "Account deleted successfully")
}