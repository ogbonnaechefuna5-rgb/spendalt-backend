package category

import (
	"github.com/gofiber/fiber/v2"
	"github.com/spendalt/backend/internal/core"
)

type Service interface {
	GetCategories() ([]*Category, error)
	GetCategoryBreakdown(userID int) ([]*CategoryBreakdown, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetCategories() ([]*Category, error) {
	return s.repo.GetAll()
}

func (s *service) GetCategoryBreakdown(userID int) ([]*CategoryBreakdown, error) {
	return s.repo.GetBreakdownByUserID(userID)
}

type Handler struct {
	core.Handler
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetCategories(c *fiber.Ctx) error {
	categories, err := h.service.GetCategories()
	if err != nil {
		return h.Fail(c, 500, err)
	}
	return h.OK(c, "categories", categories)
}

func (h *Handler) GetCategoryBreakdown(c *fiber.Ctx) error {
	breakdown, err := h.service.GetCategoryBreakdown(h.UserID(c))
	if err != nil {
		return h.Fail(c, 500, err)
	}
	return h.OK(c, "breakdown", breakdown)
}