package auth

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/spendalt/backend/internal/core"
	"github.com/spendalt/backend/internal/lang"
	"github.com/spendalt/backend/internal/monitor"
)

type Handler struct {
	core.Handler
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Signup(c *fiber.Ctx) error {
	var req struct {
		Email      string `json:"email"`
		Password   string `json:"password"`
		FirstName  string `json:"first_name"`
		MiddleName string `json:"middle_name"`
		LastName   string `json:"last_name"`
		Phone      string `json:"phone"`
	}
	if err := c.BodyParser(&req); err != nil {
		return h.Fail(c, 400, errors.New(lang.ErrInvalidBody))
	}
	user, err := h.service.Register(req.Email, req.Password, req.FirstName, req.MiddleName, req.LastName, req.Phone)
	if err != nil {
		return h.Fail(c, 400, err)
	}
	return c.Status(201).JSON(fiber.Map{
		"message": "Account created successfully",
		"user": fiber.Map{
			"id":          user.ID,
			"email":       user.Email,
			"first_name":  user.FirstName,
			"middle_name": user.MiddleName,
			"last_name":   user.LastName,
		},
	})
}

func (h *Handler) Login(c *fiber.Ctx) error {
	var req struct {
		Identifier string `json:"identifier"`
		Password   string `json:"password"`
	}
	if err := c.BodyParser(&req); err != nil {
		return h.Fail(c, 400, errors.New(lang.ErrInvalidBody))
	}
	d := monitor.ExtractDeviceInfo(c)
	user, accessToken, refreshToken, err := h.service.Login(req.Identifier, req.Password, d.DeviceID, d.IP, d.DeviceType, d.OS, d.AppVersion)
	if err != nil {
		return h.Fail(c, 401, err)
	}
	return c.JSON(fiber.Map{
		"token":         accessToken,
		"refresh_token": refreshToken,
		"user": fiber.Map{
			"id":          user.ID,
			"email":       user.Email,
			"first_name":  user.FirstName,
			"middle_name": user.MiddleName,
			"last_name":   user.LastName,
		},
	})
}

func (h *Handler) Refresh(c *fiber.Ctx) error {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.BodyParser(&req); err != nil || req.RefreshToken == "" {
		return h.Fail(c, 400, errors.New(lang.ErrRefreshRequired))
	}
	newAccess, newRefresh, err := h.service.Refresh(req.RefreshToken)
	if err != nil {
		return h.Fail(c, 401, err)
	}
	return c.JSON(fiber.Map{
		"token":         newAccess,
		"refresh_token": newRefresh,
	})
}

func (h *Handler) Logout(c *fiber.Ctx) error {
	// Revoke access token if still valid (best-effort — ignore if expired)
	if tokenString, ok := c.Locals("token").(string); ok && tokenString != "" {
		_ = h.service.Logout(tokenString)
	}
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if c.BodyParser(&req) == nil && req.RefreshToken != "" {
		_ = h.service.RevokeRefreshToken(req.RefreshToken)
	}
	return h.Message(c, "logged out successfully")
}
