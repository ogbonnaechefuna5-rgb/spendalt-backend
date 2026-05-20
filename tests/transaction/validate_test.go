package transaction_test

import (
	"testing"

	"github.com/moninte/backend/internal/core"
	"github.com/moninte/backend/internal/lang"
	"github.com/moninte/backend/internal/transaction"
	"github.com/stretchr/testify/assert"
)

type validateTxRepo struct{}

func (r *validateTxRepo) CreateRaw(raw *transaction.RawTransaction) error {
	raw.ID = "raw-1"
	return nil
}
func (r *validateTxRepo) GetRawByID(id string) (*transaction.RawTransaction, error) {
	return &transaction.RawTransaction{ID: id}, nil
}
func (r *validateTxRepo) GetUnprocessed(limit int) ([]*transaction.RawTransaction, error) { return nil, nil }
func (r *validateTxRepo) MarkProcessed(id string) error                                   { return nil }
func (r *validateTxRepo) Create(tx *transaction.Transaction) error                        { tx.ID = "tx-1"; return nil }
func (r *validateTxRepo) GetByUserID(userID string, limit, offset int) ([]*transaction.Transaction, error) {
	return nil, nil
}
func (r *validateTxRepo) GetByFingerprint(fingerprint string) (*transaction.Transaction, error) {
	return nil, core.ErrNotFound
}

func newValidateTxSvc() transaction.Service {
	repo := &validateTxRepo{}
	cat := transaction.NewRuleEngine()
	worker := transaction.NewWorker(repo, cat)
	return transaction.NewService(repo, cat, worker)
}

func TestIngestSMS_EmptyTextValidation(t *testing.T) {
	_, err := newValidateTxSvc().IngestSMS("u1", "")
	assert.EqualError(t, err, lang.ErrSMSTextRequired)
}

func TestIngestSMS_TooLong(t *testing.T) {
	_, err := newValidateTxSvc().IngestSMS("u1", string(make([]byte, 2001)))
	assert.EqualError(t, err, lang.ErrSMSTextTooLong)
}

func TestIngestManual_ZeroAmountValidation(t *testing.T) {
	_, err := newValidateTxSvc().IngestManual("u1", 0, "debit", "Shoprite", "", "")
	assert.EqualError(t, err, lang.ErrAmountRequired)
}

func TestIngestManual_NegativeAmount(t *testing.T) {
	_, err := newValidateTxSvc().IngestManual("u1", -100, "debit", "Shoprite", "", "")
	assert.EqualError(t, err, lang.ErrAmountRequired)
}

func TestIngestManual_AmountTooLarge(t *testing.T) {
	_, err := newValidateTxSvc().IngestManual("u1", 2_000_000_000, "debit", "Shoprite", "", "")
	assert.EqualError(t, err, lang.ErrAmountTooLarge)
}

func TestIngestManual_MissingType(t *testing.T) {
	_, err := newValidateTxSvc().IngestManual("u1", 1000, "", "Shoprite", "", "")
	assert.EqualError(t, err, lang.ErrTypeRequired)
}

func TestIngestManual_InvalidTypeValidation(t *testing.T) {
	_, err := newValidateTxSvc().IngestManual("u1", 1000, "transfer", "Shoprite", "", "")
	assert.EqualError(t, err, lang.ErrTypeInvalid)
}

func TestIngestManual_MissingMerchantValidation(t *testing.T) {
	_, err := newValidateTxSvc().IngestManual("u1", 1000, "debit", "", "", "")
	assert.EqualError(t, err, lang.ErrMerchantRequired)
}

func TestIngestManual_MerchantTooLong(t *testing.T) {
	_, err := newValidateTxSvc().IngestManual("u1", 1000, "debit", string(make([]byte, 101)), "", "")
	assert.EqualError(t, err, lang.ErrMerchantTooLong)
}

func TestIngestManual_DescriptionTooLong(t *testing.T) {
	_, err := newValidateTxSvc().IngestManual("u1", 1000, "debit", "Shoprite", "", string(make([]byte, 501)))
	assert.EqualError(t, err, lang.ErrDescriptionTooLong)
}
