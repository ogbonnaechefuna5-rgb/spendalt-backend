package transaction

import (
	"fmt"
	"regexp"
	"strings"
)

// TestZenithParser is exported for testing only.
func TestZenithParser(lines []string) [][]string { return parseZenithStatement(lines) }

// TestPDFParse is exported for testing only.
func TestPDFParse(data []byte) ([][]string, error) { return parsePDFBytes(data) }

// parseZenithStatement parses lines extracted from a Zenith Bank PDF statement.
// Column layout: DATE | DESCRIPTION | DEBIT | CREDIT | VALUE DATE | BALANCE
// Returns rows in [date, amount, type, merchant, reference] format for IngestCSV.
func parseZenithStatement(lines []string) [][]string {
	var rows [][]string

	// Match a line that starts with a date dd/mm/yyyy
	datePrefix := regexp.MustCompile(`^(\d{2}/\d{2}/\d{4})\s+(.+)`)
	// Match one or two money amounts at the end: debit credit [value_date] balance
	// e.g. "8.00 0.00 25/01/2026 -66.60" or "200,000.00 28/01/2026 199,933.40"
	trailingNums := regexp.MustCompile(`([\d,]+\.\d{2})\s+([\d,]+\.\d{2})\s+\d{2}/\d{2}/\d{4}\s+[-\d,]+\.\d{2}\s*$`)

	// Zenith PDFs sometimes split long descriptions across lines — join them
	var joined []string
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		if datePrefix.MatchString(line) {
			joined = append(joined, line)
		} else if len(joined) > 0 {
			// Continuation of previous line
			joined[len(joined)-1] += " " + line
		}
	}

	for _, line := range joined {
		dm := datePrefix.FindStringSubmatch(line)
		if len(dm) < 3 {
			continue
		}
		date := dm[1]
		rest := strings.TrimSpace(dm[2])

		tm := trailingNums.FindStringSubmatch(rest)
		if len(tm) < 3 {
			continue
		}

		debitStr := strings.ReplaceAll(tm[1], ",", "")
		creditStr := strings.ReplaceAll(tm[2], ",", "")

		var debit, credit float64
		fmt.Sscanf(debitStr, "%f", &debit)
		fmt.Sscanf(creditStr, "%f", &credit)

		// Skip rows where both are zero (e.g. opening balance header)
		if debit == 0 && credit == 0 {
			continue
		}

		amount := debit
		txType := "debit"
		if credit > 0 && debit == 0 {
			amount = credit
			txType = "credit"
		}

		// Description is everything before the trailing numbers
		desc := strings.TrimSpace(rest[:strings.Index(rest, tm[0])])
		if desc == "" {
			desc = "Unknown"
		}

		// Extract reference from description (long alphanumeric sequences)
		ref := ""
		if m := regexp.MustCompile(`/([A-Z0-9]{10,})`).FindStringSubmatch(desc); len(m) >= 2 {
			ref = m[1]
		}

		// Clean merchant from description — strip NIP/CIP/MOB/ prefixes
		merchant := regexp.MustCompile(`(?i)^(NIP|CIP|CR|DR|MOB|WBP|FGN|MC\s+Loc\s+POS)[/\s]+`).ReplaceAllString(desc, "")
		merchant = strings.TrimSpace(merchant)
		if len(merchant) > 60 {
			merchant = merchant[:60]
		}
		if merchant == "" {
			merchant = desc
		}

		rows = append(rows, []string{date, fmt.Sprintf("%.2f", amount), txType, merchant, ref})
	}

	return rows
}
