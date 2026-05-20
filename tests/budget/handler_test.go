package budget_test

import (
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/moninte/backend/internal/auth"
	"github.com/moninte/backend/internal/budget"
	"github.com/moninte/backend/internal/core"
	"github.com/moninte/backend/internal/lang"
	"github.com/moninte/backend/internal/testutil"
	"github.com/stretchr/testify/assert"
)

// ── stub repo ─────────────────────────────────────────────────────────────────

type stubBudgetRepo struct {
	budgets []*budget.Budget
}

func (r *stubBudgetRepo) Create(b *budget.Budget) error {
	b.ID = "budget-1"
	r.budgets = append(r.budgets, b)
	return nil
}
func (r *stubBudgetRepo) GetByUserID(userID string, limit, offset int) ([]*budget.Budget, error) {
	return r.budgets, nil
}
func (r *stubBudgetRepo) Update(b *budget.Budget) error {
	for i, existing := range r.budgets {
		if existing.ID == b.ID {
			r.budgets[i] = b
			return nil
		}
	}
	return core.ErrNotFound
}
func (r *stubBudgetRepo) Delete(id, userID string) error {
	for i, b := range r.budgets {
		if b.ID == id {
			r.budgets = append(r.budgets[:i], r.budgets[i+1:]...)
			return nil
		}
	}
	return core.ErrNotFound
}

// ── app builder ───────────────────────────────────────────────────────────────

func newBudgetApp(repo budget.Repository) *fiber.App {
	svc := budget.NewService(repo)
	h := budget.NewHandler(svc)
	app := testutil.NewApp()
	mw := auth.AuthRequired(testutil.TestSecret, auth.NewMemoryTokenStore())
	injectUser := func(c *fiber.Ctx) error {
		c.Locals("user_id", testutil.TestUserID)
		return c.Next()
	}
	app.Post("/budgets", mw, injectUser, h.CreateBudget)
	app.Get("/budgets", mw, injectUser, h.GetBudgets)
	app.Put("/budgets/:id", mw, injectUser, h.UpdateBudget)
	app.Delete("/budgets/:id", mw, injectUser, h.DeleteBudget)
	return app
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestCreateBudget_Success(t *testing.T) {
	token := testutil.MintToken(testutil.TestUserID)
	resp := testutil.Do(t, newBudgetApp(&stubBudgetRepo{}), http.MethodPost, "/budgets",
		budget.CreateBudgetRequest{Category: "Food", Amount: 50000, Period: "monthly"}, token)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var body map[string]any
	testutil.DecodeJSON(t, resp, &body)
	assert.NotNil(t, body["budget"])
}

func TestCreateBudget_MissingCategory(t *testing.T) {
	token := testutil.MintToken(testutil.TestUserID)
	resp := testutil.Do(t, newBudgetApp(&stubBudgetRepo{}), http.MethodPost, "/budgets",
		budget.CreateBudgetRequest{Amount: 50000, Period: "monthly"}, token)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	var body map[string]string
	testutil.DecodeJSON(t, resp, &body)
	assert.Equal(t, lang.ErrBudgetCategoryRequired, body["error"])
}

func TestCreateBudget_InvalidPeriod(t *testing.T) {
	token := testutil.MintToken(testutil.TestUserID)
	resp := testutil.Do(t, newBudgetApp(&stubBudgetRepo{}), http.MethodPost, "/budgets",
		budget.CreateBudgetRequest{Category: "Food", Amount: 50000, Period: "biweekly"}, token)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	var body map[string]string
	testutil.DecodeJSON(t, resp, &body)
	assert.Equal(t, lang.ErrBudgetPeriodInvalid, body["error"])
}

func TestCreateBudget_ZeroAmount(t *testing.T) {
	token := testutil.MintToken(testutil.TestUserID)
	resp := testutil.Do(t, newBudgetApp(&stubBudgetRepo{}), http.MethodPost, "/budgets",
		budget.CreateBudgetRequest{Category: "Food", Amount: 0, Period: "monthly"}, token)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	var body map[string]string
	testutil.DecodeJSON(t, resp, &body)
	assert.Equal(t, lang.ErrBudgetAmountRequired, body["error"])
}

func TestGetBudgets_Success(t *testing.T) {
	repo := &stubBudgetRepo{budgets: []*budget.Budget{
		{Category: "Food", Amount: 50000, Period: "monthly"},
	}}
	token := testutil.MintToken(testutil.TestUserID)
	resp := testutil.Do(t, newBudgetApp(repo), http.MethodGet, "/budgets", nil, token)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var body map[string]any
	testutil.DecodeJSON(t, resp, &body)
	budgets := body["budgets"].([]any)
	assert.Len(t, budgets, 1)
}

func TestGetBudgets_Empty(t *testing.T) {
	token := testutil.MintToken(testutil.TestUserID)
	resp := testutil.Do(t, newBudgetApp(&stubBudgetRepo{}), http.MethodGet, "/budgets", nil, token)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestCreateBudget_Unauthorized(t *testing.T) {
	resp := testutil.Do(t, newBudgetApp(&stubBudgetRepo{}), http.MethodPost, "/budgets",
		budget.CreateBudgetRequest{Category: "Food", Amount: 50000, Period: "monthly"}, "")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestDeleteBudget_NotFound(t *testing.T) {
	token := testutil.MintToken(testutil.TestUserID)
	resp := testutil.Do(t, newBudgetApp(&stubBudgetRepo{}), http.MethodDelete, "/budgets/nonexistent", nil, token)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
