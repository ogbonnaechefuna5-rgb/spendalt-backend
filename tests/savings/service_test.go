package savings_test

import (
	"testing"

	"github.com/moninte/backend/internal/lang"
	"github.com/moninte/backend/internal/savings"
	"github.com/stretchr/testify/assert"
)

type validateSavingsRepo struct{}

func (r *validateSavingsRepo) Create(g *savings.SavingsGoal) error                                          { return nil }
func (r *validateSavingsRepo) GetByUserID(userID string, limit, offset int) ([]*savings.SavingsGoal, error) { return nil, nil }
func (r *validateSavingsRepo) UpdateProgress(id, userID string, amount float64) error                       { return nil }
func (r *validateSavingsRepo) Delete(id, userID string) error                                               { return nil }
func (r *validateSavingsRepo) GetSummary(userID string) (float64, float64, error)                           { return 0, 0, nil }

func newValidateSavingsSvc() savings.Service {
	return savings.NewService(&validateSavingsRepo{})
}

func TestSavings_EmptyName(t *testing.T) {
	_, err := newValidateSavingsSvc().Create("u1", "", 10000, nil)
	assert.EqualError(t, err, lang.ErrGoalNameRequired)
}

func TestSavings_NameTooLong(t *testing.T) {
	_, err := newValidateSavingsSvc().Create("u1", string(make([]byte, 101)), 10000, nil)
	assert.EqualError(t, err, lang.ErrGoalNameTooLong)
}

func TestSavings_ZeroTarget(t *testing.T) {
	_, err := newValidateSavingsSvc().Create("u1", "Emergency Fund", 0, nil)
	assert.EqualError(t, err, lang.ErrTargetRequired)
}

func TestSavings_TargetTooLarge(t *testing.T) {
	_, err := newValidateSavingsSvc().Create("u1", "Emergency Fund", 2_000_000_000, nil)
	assert.EqualError(t, err, lang.ErrTargetTooLarge)
}

func TestSavings_UpdateProgressZeroAmount(t *testing.T) {
	err := newValidateSavingsSvc().UpdateProgress("id1", "u1", 0)
	assert.EqualError(t, err, lang.ErrAmountRequired)
}

func TestSavings_UpdateProgressTooLarge(t *testing.T) {
	err := newValidateSavingsSvc().UpdateProgress("id1", "u1", 2_000_000_000)
	assert.EqualError(t, err, lang.ErrAmountTooLarge)
}
