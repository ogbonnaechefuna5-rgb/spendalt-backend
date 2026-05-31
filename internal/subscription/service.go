package subscription

import "time"

type Service interface {
	GetSubscription(userID string) (*Subscription, error)
	HasEntitlement(userID, feature string) (bool, error)
	Activate(userID, planID, provider, providerRef string, periodEnd *time.Time) error
	Cancel(userID string) error
	GetPlans() ([]*Plan, error)
}

type service struct{ repo Repository }

func NewService(repo Repository) Service { return &service{repo: repo} }

func (s *service) GetSubscription(userID string) (*Subscription, error) {
	return s.repo.GetActiveByUserID(userID)
}

func (s *service) HasEntitlement(userID, feature string) (bool, error) {
	return s.repo.HasEntitlement(userID, feature)
}

func (s *service) Activate(userID, planID, provider, providerRef string, periodEnd *time.Time) error {
	return s.repo.Activate(userID, planID, provider, providerRef, periodEnd)
}

func (s *service) Cancel(userID string) error {
	return s.repo.Cancel(userID)
}

func (s *service) GetPlans() ([]*Plan, error) {
	return s.repo.GetPlans()
}
