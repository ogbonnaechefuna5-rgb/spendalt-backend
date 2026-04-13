package budget

import (
	"errors"
	"strings"
)

var validPeriods = map[string]bool{"daily": true, "weekly": true, "monthly": true, "yearly": true}

func validateBudget(category string, amount float64, period string) error {
	if strings.TrimSpace(category) == "" {
		return errors.New("category is required")
	}
	if amount <= 0 {
		return errors.New("amount must be greater than 0")
	}
	if strings.TrimSpace(period) == "" {
		return errors.New("period is required")
	}
	if !validPeriods[strings.ToLower(period)] {
		return errors.New("period must be one of: daily, weekly, monthly, yearly")
	}
	return nil
}
