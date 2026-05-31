package subscription

import "time"

type Plan struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	PriceNGN    float64   `json:"price_ngn"`
	Interval    string    `json:"interval"`
	IsActive    bool      `json:"is_active"`
	Features    []string  `json:"features,omitempty"`
}

type Subscription struct {
	ID                  string     `json:"id"`
	UserID              string     `json:"user_id"`
	PlanID              string     `json:"plan_id"`
	Status              string     `json:"status"`
	Provider            string     `json:"provider,omitempty"`
	ProviderReference   string     `json:"provider_reference,omitempty"`
	CurrentPeriodStart  time.Time  `json:"current_period_start"`
	CurrentPeriodEnd    *time.Time `json:"current_period_end,omitempty"`
	CancelledAt         *time.Time `json:"cancelled_at,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	Plan                *Plan      `json:"plan,omitempty"`
}

func (s *Subscription) IsActive() bool {
	if s.Status != "active" {
		return false
	}
	if s.CurrentPeriodEnd != nil && time.Now().After(*s.CurrentPeriodEnd) {
		return false
	}
	return true
}
