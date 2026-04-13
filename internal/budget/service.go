package budget

import "github.com/spendalt/backend/internal/core"

type Service interface {
	Create(userID, category string, amount float64, period string) (*Budget, error)
	GetByUserID(userID string, limit, offset int) ([]*Budget, error)
	Update(id, userID, category string, amount float64, period string) error
	Delete(id, userID string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(userID, category string, amount float64, period string) (*Budget, error) {
	if err := validateBudget(category, amount, period); err != nil {
		return nil, err
	}
	b := &Budget{
		UserScoped: core.UserScoped{UserID: userID},
		Category:   category,
		Amount:     amount,
		Period:     period,
	}
	return b, s.repo.Create(b)
}

func (s *service) GetByUserID(userID string, limit, offset int) ([]*Budget, error) {
	return s.repo.GetByUserID(userID, limit, offset)
}

func (s *service) Update(id, userID, category string, amount float64, period string) error {
	if err := validateBudget(category, amount, period); err != nil {
		return err
	}
	return s.repo.Update(&Budget{
		UserScoped: core.UserScoped{BaseModel: core.BaseModel{ID: id}, UserID: userID},
		Category:   category,
		Amount:     amount,
		Period:     period,
	})
}

func (s *service) Delete(id, userID string) error {
	return s.repo.Delete(id, userID)
}
