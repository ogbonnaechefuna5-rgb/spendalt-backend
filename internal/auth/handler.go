package auth

import (
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
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
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	user, err := h.service.Register(req.Email, req.Password, req.FirstName, req.MiddleName, req.LastName, req.Phone)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
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
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}
	device := c.Get("X-Device", "Unknown Device")
	ip := c.IP()
	user, accessToken, refreshToken, err := h.service.Login(req.Identifier, req.Password, device, ip)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": err.Error()})
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
		return c.Status(400).JSON(fiber.Map{"error": "refresh_token required"})
	}
	newAccess, newRefresh, err := h.service.Refresh(req.RefreshToken)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{
		"token":         newAccess,
		"refresh_token": newRefresh,
	})
}

func (h *Handler) Logout(c *fiber.Ctx) error {
	tokenString, _ := c.Locals("token").(string)
	if err := h.service.Logout(tokenString); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if c.BodyParser(&req) == nil && req.RefreshToken != "" {
		_ = h.service.RevokeRefreshToken(req.RefreshToken)
	}
	return c.JSON(fiber.Map{"message": "logged out successfully"})
}
