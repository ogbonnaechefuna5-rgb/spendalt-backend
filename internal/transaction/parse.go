package transaction

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ParseSMS parses a bank SMS alert into amount, transaction type, merchant, reference and time.
func ParseSMS(smsText string) (amount float64, txType, merchant string, err error) {
	amount, txType, merchant, _, _, _ = parseSMSFull(smsText)
	if amount == 0 {
		err = fmt.Errorf("could not parse amount from SMS")
	}
	return
}

// GenerateFingerprint produces a deterministic SHA-256 hash for deduplication.
func GenerateFingerprint(userID string, amount float64, merchant, reference string, t time.Time) string {
	return generateFingerprint(userID, amount, merchant, reference, t)
}

var (
	// Matches: NGN49,850.81 | NGN 53.75 | NGN 20000.0 | DR Amt:10,000.00 | Amt:NGN 53.75 | N3500.00
	amountPattern = regexp.MustCompile(
		`(?i)(?:` +
			`(?:DR\s*Amt|CR\s*Amt|Amt|Amount)[:\s]+(?:NGN\s*)?([\d,]+\.?\d*)` + // DR Amt:10,000.00 or Amt:NGN 53.75
			`|NGN\s*([\d,]+\.?\d*)` + // NGN49,850.81 or NGN 53.75
			`|\bN([\d,]+\.\d{2})\b` + // N3500.00 (must have decimal to avoid false matches)
		`)`)
	refPattern  = regexp.MustCompile(`(?i)(?:ref(?:erence)?[:\s#]*|tran[sx]?[:\s#]*|auth[:\s]*)([A-Z0-9]{6,30})`)
	timePattern = regexp.MustCompile(`(\d{1,2}:\d{2}(?::\d{2})?(?:\s*[AP]M)?)`)
	datePattern = regexp.MustCompile(`(\d{1,2}[\-/]\d{1,2}[\-/]\d{2,4})`)

	// SMS that are definitely not bank transactions
	noisePattern = regexp.MustCompile(`(?i)(?:data\s*plan|bundle|auto.renew|recharge\s*of|etoken|credit\s*limit\s*updated|loan\s*approved|okash|reminder|expires\s*on|dial\s*\*)`)
)

type parsedSMS struct {
	Amount    float64
	Type      string
	Merchant  string
	Reference string
	TxTime    time.Time
}

func parseSMSFull(smsText string) (amount float64, txType, merchant, reference, description string, txTime time.Time) {
	// Reject non-transaction SMS early — only if the SMS has no amount-like pattern
	if noisePattern.MatchString(smsText) && !amountPattern.MatchString(smsText) {
		return
	}

	// Amount — try each capture group in order
	matches := amountPattern.FindStringSubmatch(smsText)
	if len(matches) < 2 {
		return
	}
	// Find the first non-empty capture group (groups 1, 2, 3)
	rawAmt := ""
	for _, g := range matches[1:] {
		if g != "" {
			rawAmt = g
			break
		}
	}
	if rawAmt == "" {
		return
	}
	var err error
	amount, err = strconv.ParseFloat(strings.ReplaceAll(rawAmt, ",", ""), 64)
	if err != nil || amount <= 0 {
		return
	}

	// Type — check explicit DR/CR markers first, then keywords
	lower := strings.ToLower(smsText)
	txType = "debit"
	if regexp.MustCompile(`(?i)\b(CR\b|credit|CIP/CR|funded|received)`).MatchString(smsText) {
		txType = "credit"
	}
	if regexp.MustCompile(`(?i)\b(DR\b|debit|debited|Txn:DR)`).MatchString(smsText) {
		txType = "debit"
	}
	_ = lower

	// Merchant — try several patterns in priority order
	merchant = "Unknown"
	if m := regexp.MustCompile(`(?i)(?:transaction\s+at|debited.*?for\s+transaction\s+at|at\s+)([\w\*][\w\s\*\-\.]{1,40}?)(?:\s*\(|\s*with|\s*auth|\s*,|\s*\.|$)`).FindStringSubmatch(smsText); len(m) >= 2 {
		merchant = strings.TrimSpace(m[1])
	} else if m := regexp.MustCompile(`(?i)Des(?:cription)?[:\s]+([^\r\n,]{3,40})`).FindStringSubmatch(smsText); len(m) >= 2 {
		merchant = strings.TrimSpace(m[1])
	} else if m := regexp.MustCompile(`(?i)from\s+([A-Z][\w\s]{2,30}?)(?:\s*\(|\s*\.)`).FindStringSubmatch(smsText); len(m) >= 2 {
		merchant = strings.TrimSpace(m[1])
	}

	// Description — first meaningful non-boilerplate line that isn't the amount/balance line
	description = ""
	skipLine := regexp.MustCompile(`(?i)^(acct:|dt:|bal:|avail|dial|auto|click|reply|your\s+acct|\*\d)`)
	for _, line := range strings.Split(smsText, "\n") {
		line = strings.TrimSpace(strings.Trim(line, "\r"))
		if line == "" || skipLine.MatchString(line) {
			continue
		}
		if amountPattern.MatchString(line) {
			continue
		}
		description = line
		break
	}

	// If merchant still Unknown, derive from description
	if merchant == "Unknown" && description != "" {
		// Strip CIP/CR/MOB/ prefixes common in Zenith alerts
		clean := regexp.MustCompile(`(?i)^(CIP|CR|DR|MOB|NIP|INT|WEB)([/\s](CIP|CR|DR|MOB|NIP|INT|WEB))*[/\s]+`).ReplaceAllString(description, "")
		clean = strings.TrimSpace(clean)
		if clean != "" {
			// Take first 40 chars, trim at last space
			if len(clean) > 40 {
				clean = clean[:40]
				if i := strings.LastIndex(clean, " "); i > 10 {
					clean = clean[:i]
				}
			}
			merchant = clean
		}
	}

	// Reference number
	if m := refPattern.FindStringSubmatch(smsText); len(m) >= 2 {
		reference = strings.ToUpper(m[1])
	}

	// Time — parse from SMS text if present, combined with date if available
	txTime = time.Time{}
	timeStr := ""
	if m := timePattern.FindString(smsText); m != "" {
		timeStr = strings.TrimSpace(m)
	}
	dateStr := ""
	if m := datePattern.FindString(smsText); m != "" {
		dateStr = m
	}
	if timeStr != "" {
		formats := []string{
			"02/01/2006 15:04:05", "02-01-2006 15:04:05",
			"02/01/2006 3:04 PM", "02-01-2006 3:04 PM",
			"02/01/2006 15:04", "02-01-2006 15:04",
			"2006-01-02 15:04:05",
			"15:04:05", "15:04", "3:04 PM",
		}
		candidate := dateStr + " " + timeStr
		for _, f := range formats {
			if t, err := time.Parse(f, strings.TrimSpace(candidate)); err == nil {
				txTime = t
				break
			}
			if t, err := time.Parse(f, timeStr); err == nil {
				txTime = t
				break
			}
		}
	}
	return
}

// parseSMS is the internal version used by the worker — returns full parsed data.
func parseSMS(smsText string) (amount float64, txType, merchant string, err error) {
	amount, txType, merchant, _, _, _ = parseSMSFull(smsText)
	if amount == 0 {
		err = fmt.Errorf("could not parse amount from SMS")
	}
	return
}

// parseSMSWithMeta returns the full parsed result including reference, time and description.
func parseSMSWithMeta(smsText string) (amount float64, txType, merchant, reference, description string, txTime time.Time, err error) {
	amount, txType, merchant, reference, description, txTime = parseSMSFull(smsText)
	if amount == 0 {
		err = fmt.Errorf("could not parse amount from SMS")
	}
	return
}

// parseBalance extracts the post-transaction balance from an SMS (e.g. "Bal:NGN 1,547.05").
func parseBalance(smsText string) *float64 {
	balPat := regexp.MustCompile(`(?i)(?:Bal(?:ance)?|Avail(?:able)?\s*Bal(?:ance)?)[:\s]+(?:NGN\s*)?([\d,]+\.?\d*)`)
	m := balPat.FindStringSubmatch(smsText)
	if len(m) < 2 {
		return nil
	}
	v, err := strconv.ParseFloat(strings.ReplaceAll(m[1], ",", ""), 64)
	if err != nil {
		return nil
	}
	return &v
}
// Priority: if a reference exists, use it — it's globally unique per bank.
// Otherwise fall back to amount+merchant+time truncated to the minute.
func generateFingerprint(userID string, amount float64, merchant, reference string, t time.Time) string {
	var data string
	if reference != "" {
		// Reference-based: userID + ref is sufficient and collision-free
		data = fmt.Sprintf("%s-ref-%s", userID, reference)
	} else if !t.IsZero() {
		// Time-based: truncate to minute to handle slight clock skew
		data = fmt.Sprintf("%s-%.2f-%s-%s", userID, amount, merchant, t.Truncate(time.Minute).Format("2006-01-02T15:04"))
	} else {
		// Date-only fallback (CSV rows with no time)
		data = fmt.Sprintf("%s-%.2f-%s-%s", userID, amount, merchant, t.Format("2006-01-02"))
	}
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}
