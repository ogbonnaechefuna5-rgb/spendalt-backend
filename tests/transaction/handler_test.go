package transaction_test

import (
	"database/sql"
	"errors"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/moninte/backend/internal/auth"
	"github.com/moninte/backend/internal/core"
	"github.com/moninte/backend/internal/lang"
	"github.com/moninte/backend/internal/testutil"
	"github.com/moninte/backend/internal/transaction"
	"github.com/stretchr/testify/assert"
)

// ── stub repo ─────────────────────────────────────────────────────────────────

type stubTxRepo struct {
	txs []*transaction.Transaction
}

func (r *stubTxRepo) CreateRaw(raw *transaction.RawTransaction) error {
	raw.ID = "raw-1"
	return nil
}
func (r *stubTxRepo) GetRawByID(id string) (*transaction.RawTransaction, error) {
	return &transaction.RawTransaction{ID: id}, nil
}
func (r *stubTxRepo) GetUnprocessed(limit int) ([]*transaction.RawTransaction, error) {
	return nil, nil
}
func (r *stubTxRepo) MarkProcessed(id string) error { return nil }
func (r *stubTxRepo) Create(tx *transaction.Transaction) error {
	tx.ID = "tx-1"
	r.txs = append(r.txs, tx)
	return nil
}
func (r *stubTxRepo) GetByUserID(userID string, limit, offset int) ([]*transaction.Transaction, error) {
	return r.txs, nil
}
func (r *stubTxRepo) GetByFingerprint(fingerprint string) (*transaction.Transaction, error) {
	// Return not found so duplicates don't block tests
	return nil, core.ErrNotFound
}

// ── app builder ───────────────────────────────────────────────────────────────

func newTxApp(repo transaction.Repository) *fiber.App {
	cat := transaction.NewRuleEngine()
	worker := transaction.NewWorker(repo, cat)
	svc := transaction.NewService(repo, cat, worker)
	h := transaction.NewHandler(svc)
	app := testutil.NewApp()
	mw := auth.AuthRequired(testutil.TestSecret, auth.NewMemoryTokenStore())
	injectUser := func(c *fiber.Ctx) error {
		c.Locals("user_id", testutil.TestUserID)
		return c.Next()
	}
	app.Post("/transactions/ingest/sms", mw, injectUser, h.IngestSMS)
	app.Post("/transactions/ingest/manual", mw, injectUser, h.IngestManual)
	app.Get("/transactions", mw, injectUser, h.GetTransactions)
	return app
}

// ── ingest SMS ────────────────────────────────────────────────────────────────

func TestIngestSMS_Success(t *testing.T) {
	token := testutil.MintToken(testutil.TestUserID)
	resp := testutil.Do(t, newTxApp(&stubTxRepo{}), http.MethodPost, "/transactions/ingest/sms",
		transaction.IngestSMSRequest{SMSText: "Debit NGN 5000 at Shoprite"}, token)
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
	var body map[string]any
	testutil.DecodeJSON(t, resp, &body)
	assert.Equal(t, "SMS received and queued for processing", body["message"])
	assert.NotEmpty(t, body["id"])
}

func TestIngestSMS_EmptyText(t *testing.T) {
	token := testutil.MintToken(testutil.TestUserID)
	resp := testutil.Do(t, newTxApp(&stubTxRepo{}), http.MethodPost, "/transactions/ingest/sms",
		transaction.IngestSMSRequest{SMSText: ""}, token)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	var body map[string]string
	testutil.DecodeJSON(t, resp, &body)
	assert.Equal(t, lang.ErrSMSTextRequired, body["error"])
}

func TestIngestSMS_Unauthorized(t *testing.T) {
	resp := testutil.Do(t, newTxApp(&stubTxRepo{}), http.MethodPost, "/transactions/ingest/sms",
		transaction.IngestSMSRequest{SMSText: "NGN 5000"}, "")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// ── ingest manual ─────────────────────────────────────────────────────────────

func TestIngestManual_Success(t *testing.T) {
	token := testutil.MintToken(testutil.TestUserID)
	resp := testutil.Do(t, newTxApp(&stubTxRepo{}), http.MethodPost, "/transactions/ingest/manual",
		transaction.IngestManualRequest{Amount: 5000, Type: "debit", Merchant: "Shoprite"}, token)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var body map[string]any
	testutil.DecodeJSON(t, resp, &body)
	assert.NotNil(t, body["transaction"])
}

func TestIngestManual_ZeroAmount(t *testing.T) {
	token := testutil.MintToken(testutil.TestUserID)
	resp := testutil.Do(t, newTxApp(&stubTxRepo{}), http.MethodPost, "/transactions/ingest/manual",
		transaction.IngestManualRequest{Amount: 0, Type: "debit", Merchant: "Shoprite"}, token)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	var body map[string]string
	testutil.DecodeJSON(t, resp, &body)
	assert.Equal(t, lang.ErrAmountRequired, body["error"])
}

func TestIngestManual_InvalidType(t *testing.T) {
	token := testutil.MintToken(testutil.TestUserID)
	resp := testutil.Do(t, newTxApp(&stubTxRepo{}), http.MethodPost, "/transactions/ingest/manual",
		transaction.IngestManualRequest{Amount: 5000, Type: "transfer", Merchant: "Shoprite"}, token)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	var body map[string]string
	testutil.DecodeJSON(t, resp, &body)
	assert.Equal(t, lang.ErrTypeInvalid, body["error"])
}

func TestIngestManual_MissingMerchant(t *testing.T) {
	token := testutil.MintToken(testutil.TestUserID)
	resp := testutil.Do(t, newTxApp(&stubTxRepo{}), http.MethodPost, "/transactions/ingest/manual",
		transaction.IngestManualRequest{Amount: 5000, Type: "debit", Merchant: ""}, token)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	var body map[string]string
	testutil.DecodeJSON(t, resp, &body)
	assert.Equal(t, lang.ErrMerchantRequired, body["error"])
}

func TestIngestManual_Duplicate(t *testing.T) {
	repo := &stubTxRepo{}
	// First insert succeeds
	token := testutil.MintToken(testutil.TestUserID)
	testutil.Do(t, newTxApp(repo), http.MethodPost, "/transactions/ingest/manual",
		transaction.IngestManualRequest{Amount: 5000, Type: "debit", Merchant: "Shoprite"}, token)

	// Second identical insert should be rejected as duplicate
	dupRepo := &dupTxRepo{}
	resp := testutil.Do(t, newTxApp(dupRepo), http.MethodPost, "/transactions/ingest/manual",
		transaction.IngestManualRequest{Amount: 5000, Type: "debit", Merchant: "Shoprite"}, token)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	var body map[string]string
	testutil.DecodeJSON(t, resp, &body)
	assert.Equal(t, lang.ErrDuplicateTransaction, body["error"])
}

// ── get transactions ──────────────────────────────────────────────────────────

func TestGetTransactions_Success(t *testing.T) {
	token := testutil.MintToken(testutil.TestUserID)
	resp := testutil.Do(t, newTxApp(&stubTxRepo{}), http.MethodGet, "/transactions", nil, token)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var body map[string]any
	testutil.DecodeJSON(t, resp, &body)
	// transactions key must be present (may be null/empty for a new user)
	_, exists := body["transactions"]
	assert.True(t, exists)
}

// ── helpers ───────────────────────────────────────────────────────────────────

// dupTxRepo simulates a repo where the fingerprint already exists.
type dupTxRepo struct{ stubTxRepo }

func (r *dupTxRepo) GetByFingerprint(fingerprint string) (*transaction.Transaction, error) {
	return &transaction.Transaction{}, nil // always found = duplicate
}

var _ = errors.New
var _ = sql.ErrNoRows
