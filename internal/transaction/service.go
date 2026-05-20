package transaction

import (
	"strings"
	"time"

	"github.com/moninte/backend/internal/core"
)

type Service interface {
	IngestSMS(userID, smsText string) (*RawTransaction, error)
	IngestManual(userID string, amount float64, txType, merchant, category, description string) (*Transaction, error)
	GetTransactions(userID string, page, limit int) ([]*Transaction, error)
}

type service struct {
	repo        Repository
	categorizer Categorizer
	worker      *Worker
}

func NewService(repo Repository, categorizer Categorizer, worker *Worker) Service {
	return &service{repo: repo, categorizer: categorizer, worker: worker}
}

// IngestSMS writes the raw SMS to raw_transactions and enqueues it for async processing.
// Returns immediately so the HTTP handler can respond 202.
func (s *service) IngestSMS(userID, smsText string) (*RawTransaction, error) {
	if err := validateIngestSMS(smsText); err != nil {
		return nil, err
	}
	raw := &RawTransaction{
		UserID:  userID,
		Source:  "sms",
		RawText: smsText,
	}
	if err := s.repo.CreateRaw(raw); err != nil {
		return nil, err
	}
	s.worker.Enqueue(raw.ID)
	return raw, nil
}

// IngestManual is synchronous — the user supplied all fields, no parsing needed.
func (s *service) IngestManual(userID string, amount float64, txType, merchant, category, description string) (*Transaction, error) {
	merchant = strings.TrimSpace(merchant)
	category = strings.TrimSpace(category)
	description = strings.TrimSpace(description)
	if err := validateIngestManual(amount, txType, merchant, description); err != nil {
		return nil, err
	}
	fingerprint := generateFingerprint(userID, amount, merchant, time.Now())
	if _, err := s.repo.GetByFingerprint(fingerprint); err == nil {
		return nil, errDuplicate
	}
	tx := &Transaction{
		UserScoped:      core.UserScoped{UserID: userID},
		Amount:          amount,
		Type:            txType,
		Merchant:        merchant,
		Category:        category,
		Description:     description,
		Fingerprint:     fingerprint,
		Source:          "manual",
		TransactionDate: time.Now(),
	}
	if err := s.repo.Create(tx); err != nil {
		return nil, err
	}
	return tx, nil
}

func (s *service) GetTransactions(userID string, page, limit int) ([]*Transaction, error) {
	return s.repo.GetByUserID(userID, limit, (page-1)*limit)
}
