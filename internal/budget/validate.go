package budget

import (
	"errors"
	"strings"

	"github.com/moninte/backend/internal/lang"
)

var validPeriods = map[string]bool{"daily": true, "weekly": true, "monthly": true, "yearly": true}

func validateBudget(category string, amount float64, period string) error {
	if strings.TrimSpace(category) == "" {
		return errors.New(lang.ErrBudgetCategoryRequired)
	}
	if amount <= 0 {
		return errors.New(lang.ErrBudgetAmountRequired)
	}
	if strings.TrimSpace(period) == "" {
		return errors.New(lang.ErrBudgetPeriodRequired)
	}
	if !validPeriods[strings.ToLower(period)] {
		return errors.New(lang.ErrBudgetPeriodInvalid)
	}
	return nil
}
