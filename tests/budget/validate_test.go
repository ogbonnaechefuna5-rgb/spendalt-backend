package budget_test

import (
	"testing"

	"github.com/moninte/backend/internal/budget"
	"github.com/moninte/backend/internal/core"
	"github.com/moninte/backend/internal/lang"
	"github.com/stretchr/testify/assert"
)

type validateBudgetRepo struct{}

func (r *validateBudgetRepo) Create(b *budget.Budget) error { b.ID = "b1"; return nil }
func (r *validateBudgetRepo) GetByUserID(userID string, limit, offset int) ([]*budget.Budget, error) {
	return nil, nil
}
func (r *validateBudgetRepo) Update(b *budget.Budget) error        { return nil }
func (r *validateBudgetRepo) Delete(id, userID string) error       { return core.ErrNotFound }

func newValidateBudgetSvc() budget.Service {
	return budget.NewService(&validateBudgetRepo{})
}

func TestBudget_MissingCategory(t *testing.T) {
	_, err := newValidateBudgetSvc().Create("u1", "", 50000, "monthly")
	assert.EqualError(t, err, lang.ErrBudgetCategoryRequired)
}

func TestBudget_ZeroAmount(t *testing.T) {
	_, err := newValidateBudgetSvc().Create("u1", "Food", 0, "monthly")
	assert.EqualError(t, err, lang.ErrBudgetAmountRequired)
}

func TestBudget_NegativeAmount(t *testing.T) {
	_, err := newValidateBudgetSvc().Create("u1", "Food", -1, "monthly")
	assert.EqualError(t, err, lang.ErrBudgetAmountRequired)
}

func TestBudget_MissingPeriod(t *testing.T) {
	_, err := newValidateBudgetSvc().Create("u1", "Food", 50000, "")
	assert.EqualError(t, err, lang.ErrBudgetPeriodRequired)
}

func TestBudget_InvalidPeriod(t *testing.T) {
	_, err := newValidateBudgetSvc().Create("u1", "Food", 50000, "biweekly")
	assert.EqualError(t, err, lang.ErrBudgetPeriodInvalid)
}

func TestBudget_ValidMonthly(t *testing.T) {
	_, err := newValidateBudgetSvc().Create("u1", "Food", 50000, "monthly")
	assert.NoError(t, err)
}

func TestBudget_PeriodCaseInsensitive(t *testing.T) {
	_, err := newValidateBudgetSvc().Create("u1", "Food", 50000, "Monthly")
	assert.NoError(t, err)
}
