package user

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/moninte/backend/internal/common"
	"github.com/moninte/backend/internal/core"
)

type Handler struct {
	core.Handler
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetProfile(c *fiber.Ctx) error {
	u, err := h.service.GetProfile(h.UserID(c))
	if err != nil {
		return h.Fail(c, 500, err)
	}
	return c.JSON(ProfileResponse{User: u})
}

func (h *Handler) UpdateProfile(c *fiber.Ctx) error {
	var req UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return h.Fail(c, 400, err)
	}
	if err := h.service.UpdateProfile(h.UserID(c), req.FirstName, req.MiddleName, req.LastName, req.Phone); err != nil {
		return h.Fail(c, 400, err)
	}
	return h.Message(c, "Profile updated successfully")
}

func (h *Handler) ChangePassword(c *fiber.Ctx) error {
	var req ChangePasswordRequest
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
	return c.JSON(PreferencesResponse{Preferences: prefs})
}

func (h *Handler) SavePreferences(c *fiber.Ctx) error {
	var req SavePreferencesRequest
	if err := c.BodyParser(&req); err != nil {
		return h.Fail(c, 400, err)
	}
	p := &UserPreferences{
		SMSDetection:      req.SMSDetection,
		Analytics:         req.Analytics,
		PartnerOffers:     req.PartnerOffers,
		TransactionAlerts: req.TransactionAlerts,
		BudgetWarnings:    req.BudgetWarnings,
		AIInsights:        req.AIInsights,
		WeeklyReport:      req.WeeklyReport,
		SavingsReminders:  req.SavingsReminders,
		Promotions:        req.Promotions,
		HideBalances:      req.HideBalances,
		CrashReports:      req.CrashReports,
	}
	if err := h.service.SavePreferences(h.UserID(c), p); err != nil {
		return h.Fail(c, 500, err)
	}
	return h.Message(c, "Preferences saved")
}

func (h *Handler) GetLinkedAccounts(c *fiber.Ctx) error {
	page, limit := h.ParsePage(c)
	accounts, err := h.service.GetLinkedAccounts(h.UserID(c), limit, (page-1)*limit)
	if err != nil {
		return h.Fail(c, 500, err)
	}
	return c.JSON(LinkedAccountsResponse{
		Accounts: accounts,
		PageMeta: core.PageMeta{Page: page, Limit: limit},
	})
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
	sessions, err := h.service.GetSessions(h.UserID(c), limit, (page-1)*limit)
	if err != nil {
		return h.Fail(c, 500, err)
	}
	return c.JSON(SessionsResponse{
		Sessions: sessions,
		PageMeta: core.PageMeta{Page: page, Limit: limit},
	})
}

func (h *Handler) RevokeAllSessions(c *fiber.Ctx) error {
	if err := h.service.RevokeAllSessions(h.UserID(c)); err != nil {
		return h.Fail(c, 500, err)
	}
	return h.Message(c, "All sessions revoked")
}

func (h *Handler) RevokeSession(c *fiber.Ctx) error {
	if err := h.service.RevokeSession(h.UserID(c), c.Params(("id"))); err != nil {
		return h.Fail(c, 500, err)
	}
	return h.Message(c, "Session revoked")
}

func (h *Handler) UploadAvatar(c *fiber.Ctx) error {
	file, err := c.FormFile("avatar")
	if err != nil {
		return h.Fail(c, 400, fiber.NewError(400, "avatar file is required"))
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
		return h.Fail(c, 400, fiber.NewError(400, "only JPG, PNG and WEBP images are supported"))
	}
	if file.Size > 5*1024*1024 {
		return h.Fail(c, 400, fiber.NewError(400, "image must be under 5MB"))
	}
	if err := os.MkdirAll("./uploads/avatars", 0755); err != nil {
		return h.Fail(c, 500, err)
	}
	filename := fmt.Sprintf("%s%s", common.NewID(), ext)
	dst := filepath.Join("./uploads/avatars", filename)
	src, err := file.Open()
	if err != nil {
		return h.Fail(c, 500, err)
	}
	defer src.Close()
	out, err := os.Create(dst)
	if err != nil {
		return h.Fail(c, 500, err)
	}
	defer out.Close()
	if _, err := io.Copy(out, src); err != nil {
		return h.Fail(c, 500, err)
	}
	avatarURL := "/uploads/avatars/" + filename
	if err := h.service.UploadAvatar(h.UserID(c), avatarURL); err != nil {
		return h.Fail(c, 500, err)
	}
	return c.JSON(fiber.Map{"avatar_url": avatarURL})
}
