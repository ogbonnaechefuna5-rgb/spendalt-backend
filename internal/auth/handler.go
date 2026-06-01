package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/moninte/backend/internal/core"
	"github.com/moninte/backend/internal/lang"
	"github.com/moninte/backend/internal/monitor"
	"github.com/moninte/backend/internal/notification"
)

type Handler struct {
	core.Handler
	service      Service
	notifService notification.Service
}

func NewHandler(service Service, notifService notification.Service) *Handler {
	return &Handler{service: service, notifService: notifService}
}

func (h *Handler) Signup(c *fiber.Ctx) error {
	var req SignupRequest
	if err := c.BodyParser(&req); err != nil {
		return h.Fail(c, 400, errors.New(lang.ErrInvalidBody))
	}
	u, err := h.service.Register(req.Email, req.Password, req.FirstName, req.MiddleName, req.LastName, req.Phone)
	if err != nil {
		return h.Fail(c, 400, err)
	}
	return c.Status(fiber.StatusCreated).JSON(SignupResponse{
		Message: "Account created successfully",
		User: UserPayload{
			ID:         u.ID,
			Email:      u.Email,
			FirstName:  u.FirstName,
			MiddleName: u.MiddleName,
			LastName:   u.LastName,
		},
	})
}

func (h *Handler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return h.Fail(c, 400, errors.New(lang.ErrInvalidBody))
	}
	d := monitor.ExtractDeviceInfo(c)
	u, accessToken, refreshToken, err := h.service.Login(req.Identifier, req.Password, d.DeviceID, d.IP, d.DeviceType, d.OS, d.AppVersion)
	if err != nil {
		return h.Fail(c, 401, err)
	}

	// Send login notification asynchronously — never block the login response.
	go h.sendLoginNotification(u.ID, d.OS, d.IP)

	return c.JSON(LoginResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
		User: UserPayload{
			ID:         u.ID,
			Email:      u.Email,
			FirstName:  u.FirstName,
			MiddleName: u.MiddleName,
			LastName:   u.LastName,
		},
	})
}

func (h *Handler) Refresh(c *fiber.Ctx) error {
	var req RefreshRequest
	if err := c.BodyParser(&req); err != nil || req.RefreshToken == "" {
		return h.Fail(c, 400, errors.New(lang.ErrRefreshRequired))
	}
	newAccess, newRefresh, err := h.service.Refresh(req.RefreshToken)
	if err != nil {
		return h.Fail(c, 401, err)
	}
	return c.JSON(RefreshResponse{Token: newAccess, RefreshToken: newRefresh})
}

func (h *Handler) Logout(c *fiber.Ctx) error {
	if tokenString, ok := c.Locals("token").(string); ok && tokenString != "" {
		_ = h.service.Logout(tokenString)
	}
	var req LogoutRequest
	if c.BodyParser(&req) == nil && req.RefreshToken != "" {
		_ = h.service.RevokeRefreshToken(req.RefreshToken)
	}
	return h.Message(c, "logged out successfully")
}

func (h *Handler) OIDCLogin(c *fiber.Ctx) error {
	var req OIDCRequest
	if err := c.BodyParser(&req); err != nil || req.Provider == "" || req.IDToken == "" {
		return h.Fail(c, 400, errors.New(lang.ErrInvalidBody))
	}
	d := monitor.ExtractDeviceInfo(c)
	u, accessToken, refreshToken, err := h.service.OIDCLogin(req.Provider, req.IDToken, d.DeviceID, d.IP, d.DeviceType, d.OS, d.AppVersion)
	if err != nil {
		return h.Fail(c, 401, err)
	}

	go h.sendLoginNotification(u.ID, d.OS, d.IP)

	return c.JSON(OIDCResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
		User: UserPayload{
			ID:         u.ID,
			Email:      u.Email,
			FirstName:  u.FirstName,
			MiddleName: u.MiddleName,
			LastName:   u.LastName,
		},
	})
}

// sendLoginNotification creates an in-app + push notification for a new login.
func (h *Handler) sendLoginNotification(userID, os, ip string) {
	now := time.Now()
	body := fmt.Sprintf("New sign-in on %s at %s. If this wasn't you, secure your account immediately.",
		os, now.Format("Jan 2, 3:04 PM"))
	if ip != "" {
		body = fmt.Sprintf("New sign-in from %s on %s at %s. If this wasn't you, secure your account immediately.",
			ip, os, now.Format("Jan 2, 3:04 PM"))
	}
	_, _ = h.notifService.Send(
		context.Background(),
		userID,
		notification.TypeSystem,
		"New Login Detected",
		body,
		nil,
	)
}
