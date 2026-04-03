package budget

import (
	"strconv"
	"github.com/gofiber/fiber/v2"
	"github.com/spendalt/backend/internal/common"
	"github.com/spendalt/backend/internal/core"
)

type Budget struct {
	core.UserScoped
	Category string  `json:"category"`
	Amount   float64 `json:"amount"`
	Period   string  `json:"period"`
}

// Repository

type Repository interface {
	Create(b *Budget) error
	GetByUserID(userID int) ([]*Budget, error)
	Update(b *Budget) error
	Delete(id int) error
}

type repository struct {
	core.Repository[Budget]
}

func NewRepository(db common.DB) Repository {
	return &repository{
		core.Repository[Budget]{
			DB:    db,
			Table: "budgets",
			Scan:  func(b *Budget) []interface{} {
				return []interface{}{&b.ID, &b.UserID, &b.Category, &b.Amount, &b.Period, &b.CreatedAt}
			},
		},
	}
}

func (r *repository) Create(b *Budget) error {
	return r.DB.QueryRow(
		`INSERT INTO budgets (user_id, category, amount, period) VALUES ($1, $2, $3, $4) RETURNING id, created_at`,
		b.UserID, b.Category, b.Amount, b.Period,
	).Scan(&b.ID, &b.CreatedAt)
}

func (r *repository) GetByUserID(userID int) ([]*Budget, error) {
	return r.GetAllByUserID(userID)
}

func (r *repository) Update(b *Budget) error {
	_, err := r.DB.Exec(
		`UPDATE budgets SET category = $1, amount = $2, period = $3 WHERE id = $4`,
		b.Category, b.Amount, b.Period, b.ID,
	)
	return err
}

// Service

type Service interface {
	Create(userID int, category string, amount float64, period string) (*Budget, error)
	GetByUserID(userID int) ([]*Budget, error)
	Update(id int, category string, amount float64, period string) error
	Delete(id int) error
}

type service struct {
	core.Service[Budget]
	repo Repository
}

func NewService(repo Repository, coreRepo *core.Repository[Budget]) Service {
	return &service{
		Service: core.Service[Budget]{Repo: coreRepo},
		repo:    repo,
	}
}

func (s *service) Create(userID int, category string, amount float64, period string) (*Budget, error) {
	b := &Budget{
		UserScoped: core.UserScoped{UserID: userID},
		Category:   category,
		Amount:     amount,
		Period:     period,
	}
	return b, s.repo.Create(b)
}

func (s *service) GetByUserID(userID int) ([]*Budget, error) {
	return s.repo.GetByUserID(userID)
}

func (s *service) Update(id int, category string, amount float64, period string) error {
	return s.repo.Update(&Budget{
		UserScoped: core.UserScoped{BaseModel: core.BaseModel{ID: id}},
		Category:   category,
		Amount:     amount,
		Period:     period,
	})
}

// Handler

type Handler struct {
	core.Handler
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateBudget(c *fiber.Ctx) error {
	var req struct {
		Category string  `json:"category"`
		Amount   float64 `json:"amount"`
		Period   string  `json:"period"`
	}
	if err := c.BodyParser(&req); err != nil {
		return h.Fail(c, 400, err)
	}
	b, err := h.service.Create(h.UserID(c), req.Category, req.Amount, req.Period)
	if err != nil {
		return h.Fail(c, 400, err)
	}
	return h.Created(c, "budget", b)
}

func (h *Handler) GetBudgets(c *fiber.Ctx) error {
	budgets, err := h.service.GetByUserID(h.UserID(c))
	if err != nil {
		return h.Fail(c, 500, err)
	}
	return h.OK(c, "budgets", budgets)
}

func (h *Handler) UpdateBudget(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return h.Fail(c, 400, err)
	}
	var req struct {
		Category string  `json:"category"`
		Amount   float64 `json:"amount"`
		Period   string  `json:"period"`
	}
	if err := c.BodyParser(&req); err != nil {
		return h.Fail(c, 400, err)
	}
	if err := h.service.Update(id, req.Category, req.Amount, req.Period); err != nil {
		return h.Fail(c, 400, err)
	}
	return h.Message(c, "Budget updated successfully")
}

func (h *Handler) DeleteBudget(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return h.Fail(c, 400, err)
	}
	if err := h.service.Delete(id); err != nil {
		return h.Fail(c, 400, err)
	}
	return h.Message(c, "Budget deleted successfully")
}
