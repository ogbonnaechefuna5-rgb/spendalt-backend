package transaction

import (
	"fmt"
	"strings"
	"time"

	"github.com/moninte/backend/internal/core"
)

type Service interface {
	IngestSMS(userID, smsText string) (*RawTransaction, error)
	IngestSMSBatch(userID string, messages []string) (processed, skipped int)
	IngestManual(userID string, amount float64, txType, merchant, category, description string) (*Transaction, error)
	IngestCSV(userID string, rows [][]string) (processed, skipped int)
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
	now := time.Now()
	raw := &RawTransaction{
		UserID:     userID,
		Source:     "sms",
		RawText:    smsText,
		DetectedAt: &now,
	}
	// Pre-parse amount and type so raw_transactions columns are populated
	if amt, txType, _, _, _, _, err := parseSMSWithMeta(smsText); err == nil {
		raw.Amount = &amt
		raw.TransactionType = &txType
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
	now := time.Now()
	fingerprint := generateFingerprint(userID, amount, merchant, "", now)
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
		TransactionDate: now,
	}
	if err := s.repo.Create(tx); err != nil {
		return nil, err
	}
	return tx, nil
}

func (s *service) GetTransactions(userID string, page, limit int) ([]*Transaction, error) {
	return s.repo.GetByUserID(userID, limit, (page-1)*limit)
}

// IngestSMSBatch enqueues multiple SMS messages, skipping empty ones.
// Returns counts of queued and skipped messages.
func (s *service) IngestSMSBatch(userID string, messages []string) (processed, skipped int) {
	for _, msg := range messages {
		if _, err := s.IngestSMS(userID, msg); err != nil {
			skipped++
		} else {
			processed++
		}
	}
	return
}

// IngestCSV processes rows from a parsed CSV statement.
// Each row must have at least: date, amount, type, merchant.
// Rows that fail parsing or are duplicates are counted as skipped.
func (s *service) IngestCSV(userID string, rows [][]string) (processed, skipped int) {
	for idx, row := range rows {
		if len(row) < 4 {
			skipped++
			continue
		}
		amountStr := strings.ReplaceAll(strings.TrimSpace(row[1]), ",", "")
		var amount float64
		if _, err := fmt.Sscanf(amountStr, "%f", &amount); err != nil || validateCSVAmount(amount) != nil {
			skipped++
			continue
		}
		txType := strings.ToLower(strings.TrimSpace(row[2]))
		if txType != "debit" && txType != "credit" {
			skipped++
			continue
		}
		merchant := strings.TrimSpace(row[3])
		if merchant == "" {
			skipped++
			continue
		}
		txTime := parseRowDateTime(strings.TrimSpace(row[0]))
		reference := ""
		if len(row) >= 5 {
			reference = strings.TrimSpace(row[4])
		}
		// Use row index as tiebreaker when no explicit reference exists,
		// preventing same-day same-amount rows from colliding on fingerprint
		if reference == "" {
			reference = fmt.Sprintf("upload-row-%d", idx)
		}
		fingerprint := generateFingerprint(userID, amount, merchant, reference, txTime)
		if _, err := s.repo.GetByFingerprint(fingerprint); err == nil {
			skipped++ // duplicate
			continue
		}
		category := s.categorizer.Categorize(merchant)
		tx := &Transaction{
			UserScoped:      core.UserScoped{UserID: userID},
			Amount:          amount,
			Type:            txType,
			Merchant:        merchant,
			Category:        category,
			Fingerprint:     fingerprint,
			Source:          "upload",
			TransactionDate: txTime,
		}
		if err := s.repo.Create(tx); err != nil {
			skipped++
			continue
		}
		processed++
	}
	return
}

// parseRowDateTime tries common date/datetime formats used by Nigerian banks.
func parseRowDateTime(s string) time.Time {
	if s == "" {
		return time.Now()
	}
	formats := []string{
		"02/01/2006 15:04:05",
		"02-01-2006 15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
		"02/01/2006 15:04",
		"02-01-2006 15:04",
		"2006-01-02 15:04",
		"02/01/2006",
		"02-01-2006",
		"2006-01-02",
		"Jan 2, 2006",
		"2 Jan 2006",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t
		}
	}
	return time.Now()
}
