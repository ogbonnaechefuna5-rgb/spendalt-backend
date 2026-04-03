package transaction

import (
	"strconv"
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

func (h *Handler) IngestSMS(c *fiber.Ctx) error {
	var req struct {
		SMSText string `json:"sms_text"`
	}
	if err := c.BodyParser(&req); err != nil {
		return h.Fail(c, 400, err)
	}
	tx, err := h.service.IngestSMS(h.UserID(c), req.SMSText)
	if err != nil {
		return h.Fail(c, 400, err)
	}
	return h.Created(c, "transaction", tx)
}

func (h *Handler) IngestManual(c *fiber.Ctx) error {
	var req struct {
		Amount      float64 `json:"amount"`
		Type        string  `json:"type"`
		Merchant    string  `json:"merchant"`
		Category    string  `json:"category"`
		Description string  `json:"description"`
	}
	if err := c.BodyParser(&req); err != nil {
		return h.Fail(c, 400, err)
	}
	tx, err := h.service.IngestManual(h.UserID(c), req.Amount, req.Type, req.Merchant, req.Category, req.Description)
	if err != nil {
		return h.Fail(c, 400, err)
	}
	return h.Created(c, "transaction", tx)
}

func (h *Handler) GetTransactions(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	transactions, err := h.service.GetTransactions(h.UserID(c), page, limit)
	if err != nil {
		return h.Fail(c, 500, err)
	}
	return c.JSON(fiber.Map{"transactions": transactions, "page": page, "limit": limit})
}