package transaction

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"github.com/spendalt/backend/internal/core"
)

type Service interface {
	IngestSMS(userID int, smsText string) (*Transaction, error)
	IngestManual(userID int, amount float64, txType, merchant, category, description string) (*Transaction, error)
	GetTransactions(userID int, page, limit int) ([]*Transaction, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) IngestSMS(userID int, smsText string) (*Transaction, error) {
	amount, txType, merchant, err := s.parseSMS(smsText)
	if err != nil {
		return nil, err
	}

	fingerprint := s.generateFingerprint(userID, amount, merchant, time.Now())

	if _, err := s.repo.GetByFingerprint(fingerprint); err == nil {
		return nil, fmt.Errorf("duplicate transaction")
	}

	category := s.categorizeTransaction(merchant)

	tx := &Transaction{
		UserScoped:      core.UserScoped{UserID: userID},
		Amount:          amount,
		Type:            txType,
		Merchant:        merchant,
		Category:        category,
		Fingerprint:     fingerprint,
		TransactionDate: time.Now(),
	}

	if err := s.repo.Create(tx); err != nil {
		return nil, err
	}

	return tx, nil
}

func (s *service) IngestManual(userID int, amount float64, txType, merchant, category, description string) (*Transaction, error) {
	fingerprint := s.generateFingerprint(userID, amount, merchant, time.Now())

	tx := &Transaction{
		UserScoped:      core.UserScoped{UserID: userID},
		Amount:          amount,
		Type:            txType,
		Merchant:        merchant,
		Category:        category,
		Description:     description,
		Fingerprint:     fingerprint,
		TransactionDate: time.Now(),
	}

	if err := s.repo.Create(tx); err != nil {
		return nil, err
	}

	return tx, nil
}

func (s *service) GetTransactions(userID int, page, limit int) ([]*Transaction, error) {
	offset := (page - 1) * limit
	return s.repo.GetByUserID(userID, limit, offset)
}

func (s *service) parseSMS(smsText string) (float64, string, string, error) {
	amountRegex := regexp.MustCompile(`NGN\s*([\d,]+\.?\d*)`)
	matches := amountRegex.FindStringSubmatch(smsText)
	if len(matches) < 2 {
		return 0, "", "", fmt.Errorf("could not parse amount")
	}

	amountStr := strings.ReplaceAll(matches[1], ",", "")
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return 0, "", "", err
	}

	txType := "debit"
	if strings.Contains(strings.ToLower(smsText), "credit") {
		txType = "credit"
	}

	merchant := "Unknown"
	if strings.Contains(smsText, "at ") {
		parts := strings.Split(smsText, "at ")
		if len(parts) > 1 {
			merchant = strings.TrimSpace(strings.Split(parts[1], " ")[0])
		}
	}

	return amount, txType, merchant, nil
}

func (s *service) generateFingerprint(userID int, amount float64, merchant string, date time.Time) string {
	data := fmt.Sprintf("%d-%.2f-%s-%s", userID, amount, merchant, date.Format("2006-01-02"))
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

func (s *service) categorizeTransaction(merchant string) string {
	merchant = strings.ToLower(strings.TrimSpace(merchant))
	
	categories := map[string][]string{
		"Food & Dining": {"restaurant", "food", "cafe", "pizza", "burger", "kfc", "dominos"},
		"Transportation": {"uber", "bolt", "fuel", "petrol", "transport", "bus", "taxi"},
		"Shopping": {"mall", "store", "shop", "market", "supermarket", "shoprite", "jumia"},
		"Entertainment": {"cinema", "movie", "game", "club", "bar", "netflix", "spotify"},
		"Utilities": {"electric", "water", "internet", "phone", "bill", "mtn", "airtel"},
	}

	for category, keywords := range categories {
		for _, keyword := range keywords {
			if strings.Contains(merchant, keyword) {
				return category
			}
		}
	}

	return "Other"
}