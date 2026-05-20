package transaction

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ParseSMS parses a bank SMS alert into amount, transaction type, and merchant.
func ParseSMS(smsText string) (amount float64, txType, merchant string, err error) {
	return parseSMS(smsText)
}

// GenerateFingerprint produces a deterministic SHA-256 hash for deduplication.
func GenerateFingerprint(userID string, amount float64, merchant string, date time.Time) string {
	return generateFingerprint(userID, amount, merchant, date)
}

func parseSMS(smsText string) (amount float64, txType, merchant string, err error) {
	matches := regexp.MustCompile(`NGN\s*([\d,]+\.?\d*)`).FindStringSubmatch(smsText)
	if len(matches) < 2 {
		return 0, "", "", fmt.Errorf("could not parse amount from SMS")
	}
	amount, err = strconv.ParseFloat(strings.ReplaceAll(matches[1], ",", ""), 64)
	if err != nil {
		return
	}
	txType = "debit"
	if strings.Contains(strings.ToLower(smsText), "credit") {
		txType = "credit"
	}
	merchant = "Unknown"
	if parts := strings.SplitN(smsText, "at ", 2); len(parts) == 2 {
		if fields := strings.Fields(parts[1]); len(fields) > 0 {
			merchant = strings.TrimSpace(fields[0])
		}
	}
	return
}

func generateFingerprint(userID string, amount float64, merchant string, date time.Time) string {
	data := fmt.Sprintf("%s-%.2f-%s-%s", userID, amount, merchant, date.Format("2006-01-02"))
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}
