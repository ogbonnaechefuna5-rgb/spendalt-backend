package transaction

import (
	"errors"
	"strings"
)

var errDuplicate = errors.New("duplicate transaction")

var validTypes = map[string]bool{"debit": true, "credit": true}

func validateIngestSMS(smsText string) error {
	if strings.TrimSpace(smsText) == "" {
		return errors.New("sms_text is required")
	}
	if len(smsText) > 2000 {
		return errors.New("sms_text is too long")
	}
	return nil
}

func validateIngestManual(amount float64, txType, merchant, description string) error {
	if amount <= 0 {
		return errors.New("amount must be greater than 0")
	}
	if amount > 1_000_000_000 {
		return errors.New("amount exceeds maximum allowed value")
	}
	if strings.TrimSpace(txType) == "" {
		return errors.New("type is required")
	}
	if !validTypes[strings.ToLower(txType)] {
		return errors.New("type must be 'debit' or 'credit'")
	}
	if strings.TrimSpace(merchant) == "" {
		return errors.New("merchant is required")
	}
	if len(merchant) > 100 {
		return errors.New("merchant name is too long")
	}
	if len(description) > 500 {
		return errors.New("description is too long")
	}
	return nil
}
