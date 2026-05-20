package savings

import (
	"errors"
	"strings"
	"time"

	"github.com/moninte/backend/internal/lang"
)

type Service interface {
	Create(userID, name string, target float64, deadline *time.Time) (*SavingsGoal, error)
	GetByUserID(userID string, limit, offset int) ([]*SavingsGoal, error)
	UpdateProgress(id, userID string, amount float64) error
	Delete(id, userID string) error
	GetComposite(userID string, limit, offset int) (*SavingsResponse, error)
}

type service struct{ repo Repository }

func NewService(repo Repository) Service { return &service{repo: repo} }

func (s *service) Create(userID, name string, target float64, deadline *time.Time) (*SavingsGoal, error) {
	if strings.TrimSpace(name) == "" {
		return nil, errors.New(lang.ErrGoalNameRequired)
	}
	if len(name) > 100 {
		return nil, errors.New(lang.ErrGoalNameTooLong)
	}
	if target <= 0 {
		return nil, errors.New(lang.ErrTargetRequired)
	}
	if target > 1_000_000_000 {
		return nil, errors.New(lang.ErrTargetTooLarge)
	}
	g := &SavingsGoal{UserID: userID, Name: strings.TrimSpace(name), TargetAmount: target, Deadline: deadline}
	return g, s.repo.Create(g)
}

func (s *service) GetByUserID(userID string, limit, offset int) ([]*SavingsGoal, error) {
	return s.repo.GetByUserID(userID, limit, offset)
}

func (s *service) UpdateProgress(id, userID string, amount float64) error {
	if amount <= 0 {
		return errors.New(lang.ErrAmountRequired)
	}
	if amount > 1_000_000_000 {
		return errors.New(lang.ErrAmountTooLarge)
	}
	return s.repo.UpdateProgress(id, userID, amount)
}

func (s *service) Delete(id, userID string) error {
	return s.repo.Delete(id, userID)
}

func (s *service) GetComposite(userID string, limit, offset int) (*SavingsResponse, error) {
	totalSaved, monthlyGain, err := s.repo.GetSummary(userID)
	if err != nil {
		return nil, err
	}
	goals, err := s.repo.GetByUserID(userID, limit, offset)
	if err != nil {
		return nil, err
	}
	return &SavingsResponse{
		TotalSaved:  totalSaved,
		MonthlyGain: monthlyGain,
		Goals:       goals,
	}, nil
}
