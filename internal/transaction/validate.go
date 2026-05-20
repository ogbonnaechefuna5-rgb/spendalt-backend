package transaction

import (
	"errors"
	"strings"

	"github.com/moninte/backend/internal/lang"
)

var errDuplicate = errors.New(lang.ErrDuplicateTransaction)

var validTypes = map[string]bool{"debit": true, "credit": true}

func validateIngestSMS(smsText string) error {
	if strings.TrimSpace(smsText) == "" {
		return errors.New(lang.ErrSMSTextRequired)
	}
	if len(smsText) > 2000 {
		return errors.New(lang.ErrSMSTextTooLong)
	}
	return nil
}

func validateIngestManual(amount float64, txType, merchant, description string) error {
	if amount <= 0 {
		return errors.New(lang.ErrAmountRequired)
	}
	if amount > 1_000_000_000 {
		return errors.New(lang.ErrAmountTooLarge)
	}
	if strings.TrimSpace(txType) == "" {
		return errors.New(lang.ErrTypeRequired)
	}
	if !validTypes[strings.ToLower(txType)] {
		return errors.New(lang.ErrTypeInvalid)
	}
	if strings.TrimSpace(merchant) == "" {
		return errors.New(lang.ErrMerchantRequired)
	}
	if len(merchant) > 100 {
		return errors.New(lang.ErrMerchantTooLong)
	}
	if len(description) > 500 {
		return errors.New(lang.ErrDescriptionTooLong)
	}
	return nil
}
