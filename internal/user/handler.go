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
		FirstName  string `json:"first_name"`
		MiddleName string `json:"middle_name"`
		LastName   string `json:"last_name"`
		Phone      string `json:"phone"`
	}
	if err := c.BodyParser(&req); err != nil {
		return h.Fail(c, 400, err)
	}
	if err := h.service.UpdateProfile(h.UserID(c), req.FirstName, req.MiddleName, req.LastName, req.Phone); err != nil {
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

func (h *Handler) GetPreferences(c *fiber.Ctx) error {
	prefs, err := h.service.GetPreferences(h.UserID(c))
	if err != nil {
		return h.Fail(c, 500, err)
	}
	return h.OK(c, "preferences", prefs)
}

func (h *Handler) SavePreferences(c *fiber.Ctx) error {
	var req struct {
		SMSDetection  bool `json:"sms_detection"`
		Analytics     bool `json:"analytics"`
		PartnerOffers bool `json:"partner_offers"`
	}
	if err := c.BodyParser(&req); err != nil {
		return h.Fail(c, 400, err)
	}
	if err := h.service.SavePreferences(h.UserID(c), req.SMSDetection, req.Analytics, req.PartnerOffers); err != nil {
		return h.Fail(c, 500, err)
	}
	return h.Message(c, "Preferences saved")
}

func (h *Handler) GetLinkedAccounts(c *fiber.Ctx) error {
	page, limit := h.ParsePage(c)
	offset := (page - 1) * limit
	accounts, err := h.service.GetLinkedAccounts(h.UserID(c), limit, offset)
	if err != nil {
		return h.Fail(c, 500, err)
	}
	return c.JSON(fiber.Map{"accounts": accounts, "page": page, "limit": limit})
}

func (h *Handler) RemoveLinkedAccount(c *fiber.Ctx) error {
	if err := h.service.RemoveLinkedAccount(h.UserID(c), c.Params("id")); err != nil {
		return h.Fail(c, 400, err)
	}
	return h.Message(c, "Account removed")
}

func (h *Handler) SyncLinkedAccount(c *fiber.Ctx) error {
	if err := h.service.SyncLinkedAccount(h.UserID(c), c.Params("id")); err != nil {
		return h.Fail(c, 400, err)
	}
	return h.Message(c, "Account synced")
}

func (h *Handler) GetSessions(c *fiber.Ctx) error {
	page, limit := h.ParsePage(c)
	offset := (page - 1) * limit
	sessions, err := h.service.GetSessions(h.UserID(c), limit, offset)
	if err != nil {
		return h.Fail(c, 500, err)
	}
	return c.JSON(fiber.Map{"sessions": sessions, "page": page, "limit": limit})
}

func (h *Handler) RevokeAllSessions(c *fiber.Ctx) error {
	if err := h.service.RevokeAllSessions(h.UserID(c)); err != nil {
		return h.Fail(c, 500, err)
	}
	return h.Message(c, "All sessions revoked")
}