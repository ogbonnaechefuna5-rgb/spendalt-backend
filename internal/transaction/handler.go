package transaction

import (
	"github.com/gofiber/fiber/v2"
	"github.com/moninte/backend/internal/core"
)

type Handler struct {
	core.Handler
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) IngestSMS(c *fiber.Ctx) error {
	var req IngestSMSRequest
	if err := c.BodyParser(&req); err != nil {
		return h.Fail(c, 400, err)
	}
	raw, err := h.service.IngestSMS(h.UserID(c), req.SMSText)
	if err != nil {
		return h.Fail(c, 400, err)
	}
	return c.Status(fiber.StatusAccepted).JSON(IngestSMSResponse{
		Message: "SMS received and queued for processing",
		ID:      raw.ID,
	})
}

func (h *Handler) IngestManual(c *fiber.Ctx) error {
	var req IngestManualRequest
	if err := c.BodyParser(&req); err != nil {
		return h.Fail(c, 400, err)
	}
	tx, err := h.service.IngestManual(h.UserID(c), req.Amount, req.Type, req.Merchant, req.Category, req.Description)
	if err != nil {
		return h.Fail(c, 400, err)
	}
	return c.Status(fiber.StatusCreated).JSON(TransactionResponse{Transaction: tx})
}

func (h *Handler) GetTransactions(c *fiber.Ctx) error {
	page, limit := h.ParsePage(c)
	txs, err := h.service.GetTransactions(h.UserID(c), page, limit)
	if err != nil {
		return h.Fail(c, 500, err)
	}
	return c.JSON(TransactionListResponse{
		Transactions: txs,
		PageMeta:     core.PageMeta{Page: page, Limit: limit},
	})
}
