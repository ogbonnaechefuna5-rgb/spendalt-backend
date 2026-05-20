package auth

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/moninte/backend/internal/core"
	"github.com/moninte/backend/internal/lang"
	"github.com/moninte/backend/internal/monitor"
)

type Handler struct {
	core.Handler
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
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
