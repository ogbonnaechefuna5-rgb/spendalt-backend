package savings

import (
	"errors"
	"strings"
	"time"
)

type Service interface {
	Create(userID, name string, target float64, deadline *time.Time) (*SavingsGoal, error)
	GetByUserID(userID string, limit, offset int) ([]*SavingsGoal, error)
	UpdateProgress(id, userID string, amount float64) error
	Delete(id, userID string) error
}

type service struct{ repo Repository }

func NewService(repo Repository) Service { return &service{repo: repo} }

func (s *service) Create(userID, name string, target float64, deadline *time.Time) (*SavingsGoal, error) {
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("goal name is required")
	}
	if len(name) > 100 {
		return nil, errors.New("goal name is too long")
	}
	if target <= 0 {
		return nil, errors.New("target amount must be greater than 0")
	}
	if target > 1_000_000_000 {
		return nil, errors.New("target amount exceeds maximum allowed value")
	}
	g := &SavingsGoal{UserID: userID, Name: strings.TrimSpace(name), TargetAmount: target, Deadline: deadline}
	return g, s.repo.Create(g)
}

func (s *service) GetByUserID(userID string, limit, offset int) ([]*SavingsGoal, error) {
	return s.repo.GetByUserID(userID, limit, offset)
}

func (s *service) UpdateProgress(id, userID string, amount float64) error {
	if amount <= 0 {
		return errors.New("amount must be greater than 0")
	}
	if amount > 1_000_000_000 {
		return errors.New("amount exceeds maximum allowed value")
	}
	return s.repo.UpdateProgress(id, userID, amount)
}

func (s *service) Delete(id, userID string) error {
	return s.repo.Delete(id, userID)
}
