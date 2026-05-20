package savings_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/moninte/backend/internal/auth"
	"github.com/moninte/backend/internal/core"
	"github.com/moninte/backend/internal/lang"
	"github.com/moninte/backend/internal/savings"
	"github.com/moninte/backend/internal/testutil"
	"github.com/stretchr/testify/assert"
)

// ── stub repo ─────────────────────────────────────────────────────────────────

type stubSavingsRepo struct {
	goals []*savings.SavingsGoal
}

func (r *stubSavingsRepo) Create(g *savings.SavingsGoal) error {
	g.ID = "goal-1"
	r.goals = append(r.goals, g)
	return nil
}
func (r *stubSavingsRepo) GetByUserID(userID string, limit, offset int) ([]*savings.SavingsGoal, error) {
	return r.goals, nil
}
func (r *stubSavingsRepo) UpdateProgress(id, userID string, amount float64) error {
	for _, g := range r.goals {
		if g.ID == id {
			return nil
		}
	}
	return core.ErrNotFound
}
func (r *stubSavingsRepo) Delete(id, userID string) error {
	for i, g := range r.goals {
		if g.ID == id {
			r.goals = append(r.goals[:i], r.goals[i+1:]...)
			return nil
		}
	}
	return core.ErrNotFound
}
func (r *stubSavingsRepo) GetSummary(userID string) (float64, float64, error) {
	var total float64
	for _, g := range r.goals {
		total += g.CurrentAmount
	}
	return total, 0, nil
}

// ── app builder ───────────────────────────────────────────────────────────────

func newSavingsApp(repo savings.Repository) *fiber.App {
	svc := savings.NewService(repo)
	h := savings.NewHandler(svc)
	app := testutil.NewApp()
	mw := auth.AuthRequired(testutil.TestSecret, auth.NewMemoryTokenStore())
	injectUser := func(c *fiber.Ctx) error {
		c.Locals("user_id", testutil.TestUserID)
		return c.Next()
	}
	app.Post("/savings", mw, injectUser, h.CreateGoal)
	app.Get("/savings", mw, injectUser, h.GetComposite)
	app.Put("/savings/:id/progress", mw, injectUser, h.UpdateProgress)
	app.Delete("/savings/:id", mw, injectUser, h.DeleteGoal)
	return app
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestCreateGoal_Success(t *testing.T) {
	token := testutil.MintToken(testutil.TestUserID)
	resp := testutil.Do(t, newSavingsApp(&stubSavingsRepo{}), http.MethodPost, "/savings",
		savings.CreateGoalRequest{Name: "Emergency Fund", TargetAmount: 500000}, token)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var body map[string]any
	testutil.DecodeJSON(t, resp, &body)
	assert.NotNil(t, body["goal"])
}

func TestCreateGoal_EmptyName(t *testing.T) {
	token := testutil.MintToken(testutil.TestUserID)
	resp := testutil.Do(t, newSavingsApp(&stubSavingsRepo{}), http.MethodPost, "/savings",
		savings.CreateGoalRequest{Name: "", TargetAmount: 500000}, token)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	var body map[string]string
	testutil.DecodeJSON(t, resp, &body)
	assert.Equal(t, lang.ErrGoalNameRequired, body["error"])
}

func TestCreateGoal_ZeroTarget(t *testing.T) {
	token := testutil.MintToken(testutil.TestUserID)
	resp := testutil.Do(t, newSavingsApp(&stubSavingsRepo{}), http.MethodPost, "/savings",
		savings.CreateGoalRequest{Name: "Fund", TargetAmount: 0}, token)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	var body map[string]string
	testutil.DecodeJSON(t, resp, &body)
	assert.Equal(t, lang.ErrTargetRequired, body["error"])
}

func TestGetSavings_Success(t *testing.T) {
	repo := &stubSavingsRepo{goals: []*savings.SavingsGoal{
		{ID: "g1", Name: "Emergency Fund", TargetAmount: 500000, CurrentAmount: 100000, Status: "active"},
	}}
	token := testutil.MintToken(testutil.TestUserID)
	resp := testutil.Do(t, newSavingsApp(repo), http.MethodGet, "/savings", nil, token)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var body map[string]any
	testutil.DecodeJSON(t, resp, &body)
	assert.NotNil(t, body["goals"])
	assert.Equal(t, float64(100000), body["totalSaved"])
}

func TestUpdateProgress_Success(t *testing.T) {
	repo := &stubSavingsRepo{goals: []*savings.SavingsGoal{
		{ID: "goal-1", Name: "Fund", TargetAmount: 500000},
	}}
	token := testutil.MintToken(testutil.TestUserID)
	resp := testutil.Do(t, newSavingsApp(repo), http.MethodPut, "/savings/goal-1/progress",
		savings.UpdateProgressRequest{Amount: 10000}, token)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestUpdateProgress_ZeroAmount(t *testing.T) {
	token := testutil.MintToken(testutil.TestUserID)
	resp := testutil.Do(t, newSavingsApp(&stubSavingsRepo{}), http.MethodPut, "/savings/goal-1/progress",
		savings.UpdateProgressRequest{Amount: 0}, token)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	var body map[string]string
	testutil.DecodeJSON(t, resp, &body)
	assert.Equal(t, lang.ErrAmountRequired, body["error"])
}

func TestSavings_Unauthorized(t *testing.T) {
	resp := testutil.Do(t, newSavingsApp(&stubSavingsRepo{}), http.MethodGet, "/savings", nil, "")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestDeleteGoal_NotFound(t *testing.T) {
	token := testutil.MintToken(testutil.TestUserID)
	resp := testutil.Do(t, newSavingsApp(&stubSavingsRepo{}), http.MethodDelete, "/savings/nonexistent", nil, token)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

var _ = time.Now
