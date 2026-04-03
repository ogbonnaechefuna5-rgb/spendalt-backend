package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GenerateFingerprint(userID string, amount float64, date time.Time, merchant string) string {
	data := fmt.Sprintf("%s:%.2f:%s:%s", userID, amount, date.Format("2006-01-02"), merchant)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func NormalizeMerchant(name string) string {
	name = strings.ToLower(name)
	name = regexp.MustCompile(`[^a-z0-9\s]`).ReplaceAllString(name, "")
	name = strings.TrimSpace(name)
	return name
}

func ExtractAmountFromSMS(text string) (float64, string) {
	text = strings.ToLower(text)
	
	// Detect debit/credit
	txType := "debit"
	if strings.Contains(text, "credit") || strings.Contains(text, "received") {
		txType = "credit"
	}
	
	// Extract amount (NGN format)
	re := regexp.MustCompile(`(?:ngn|₦|n)\s*([0-9,]+\.?[0-9]*)`)
	matches := re.FindStringSubmatch(text)
	
	if len(matches) > 1 {
		amountStr := strings.ReplaceAll(matches[1], ",", "")
		var amount float64
		fmt.Sscanf(amountStr, "%f", &amount)
		return amount, txType
	}
	
	return 0, txType
}

func CategorizeTransaction(description string, categories []string, keywords map[string][]string) string {
	desc := strings.ToLower(description)
	
	for category, kws := range keywords {
		for _, kw := range kws {
			if strings.Contains(desc, strings.ToLower(kw)) {
				return category
			}
		}
	}
	
	return "Other"
}
