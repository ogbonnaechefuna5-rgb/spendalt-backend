package subscription

import (
	"github.com/gofiber/fiber/v2"
	"github.com/moninte/backend/internal/core"
	"github.com/moninte/backend/internal/payment"
)

type Handler struct {
	core.Handler
	service  Service
	payments *payment.Registry
}

func NewHandler(service Service, payments *payment.Registry) *Handler {
	return &Handler{service: service, payments: payments}
}

func (h *Handler) GetPlans(c *fiber.Ctx) error {
	plans, err := h.service.GetPlans()
	if err != nil {
		return h.Fail(c, 500, err)
	}
	return c.JSON(fiber.Map{"plans": plans})
}

func (h *Handler) GetSubscription(c *fiber.Ctx) error {
	sub, err := h.service.GetSubscription(h.UserID(c))
	if err != nil {
		return h.Fail(c, 500, err)
	}
	return c.JSON(fiber.Map{"subscription": sub})
}

func (h *Handler) Cancel(c *fiber.Ctx) error {
	if err := h.service.Cancel(h.UserID(c)); err != nil {
		return h.Fail(c, 400, err)
	}
	return h.Message(c, "Subscription cancelled")
}

// Webhook is the single entry point for all payment providers.
// POST /webhooks/:provider
func (h *Handler) Webhook(c *fiber.Ctx) error {
	providerName := c.Params("provider")
	provider, ok := h.payments.Get(providerName)
	if !ok {
		return c.Status(400).JSON(fiber.Map{"error": "unknown provider"})
	}

	event, err := provider.HandleWebhook(c.Body(), c.Get("X-Paystack-Signature"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	if event.Type == payment.EventChargeSuccess {
		if err := h.service.Activate(
			event.UserID, event.PlanID, providerName,
			event.Reference, event.PeriodEnd,
		); err != nil {
			return h.Fail(c, 500, err)
		}
	}

	return c.SendStatus(200)
}
