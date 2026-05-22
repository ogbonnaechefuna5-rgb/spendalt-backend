package transaction

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/ledongthuc/pdf"
)

// parsePDFBytes is the entry point used by the handler.
func parsePDFBytes(data []byte) ([][]string, error) {
	ra := &pdfReaderAt{b: data}
	return parsePDF(ra, int64(len(data)))
}

func parsePDF(r io.ReaderAt, size int64) ([][]string, error) {
	reader, err := pdf.NewReader(r, size)
	if err != nil {
		return nil, fmt.Errorf("could not open PDF: %w", err)
	}

	var sb strings.Builder
	for i := 1; i <= reader.NumPage(); i++ {
		page := reader.Page(i)
		if page.V.IsNull() {
			continue
		}
		text, err := page.GetPlainText(nil)
		if err != nil {
			continue
		}
		sb.WriteString(text)
		sb.WriteString("\n")
	}

	lines := splitLines(sb.String())
	if len(lines) == 0 {
		return nil, fmt.Errorf("no text found in PDF")
	}

	// Try formats in order of specificity
	if rows := parseZenithStatement(lines); len(rows) > 0 {
		return rows, nil
	}
	if rows := extractOPayRows(lines); len(rows) > 0 {
		return rows, nil
	}
	if rows := extractSplitAmountRows(lines); len(rows) > 0 {
		return rows, nil
	}
	if rows := extractTabularRows(lines); len(rows) > 0 {
		return rows, nil
	}
	rows := extractSMSRows(lines)
	return rows, nil
}

// ── Regexes ───────────────────────────────────────────────────────────────────

var (
	amountRe = regexp.MustCompile(`[\d,]+\.\d{2}`)
	dateRe   = regexp.MustCompile(`\d{1,2}[/\-]\d{1,2}[/\-]\d{2,4}`)
	typeRe   = regexp.MustCompile(`(?i)\b(DR|CR|debit|credit)\b`)

	// OPay: "21 Mar 2026 08:15:43" or "21 Mar 2026"
	opayDateRe = regexp.MustCompile(`(?i)\d{1,2}\s+(?:Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)\s+\d{4}`)
)

// ── OPay layout ───────────────────────────────────────────────────────────────
// Columns: Trans.Time | Value Date | Description | Debit(₦) | Credit(₦) | Balance After(₦) | Channel | Ref
// Empty amount cells contain "--" not "0.00".
// PDF text extraction merges multi-line descriptions onto one logical row,
// so we join lines until we see the next date anchor.

func extractOPayRows(lines []string) [][]string {
	// Merge continuation lines: a new transaction starts when a line begins
	// with an OPay-style date ("21 Mar 2026 ...").
	var merged []string
	for _, l := range lines {
		if opayDateRe.MatchString(l) && (len(l) > 0 && (l[0] >= '0' && l[0] <= '9')) {
			merged = append(merged, l)
		} else if len(merged) > 0 {
			merged[len(merged)-1] += " " + l
		}
	}

	var rows [][]string
	for _, line := range merged {
		// Must contain an OPay date
		if !opayDateRe.MatchString(line) {
			continue
		}

		// Find all numeric amounts (ignore "--" placeholders)
		amounts := amountRe.FindAllString(line, -1)
		if len(amounts) < 2 {
			// Need at least txn amount + balance
			continue
		}

		// Last amount = Balance After — drop it
		txAmounts := amounts[:len(amounts)-1]

		// Determine debit vs credit.
		// OPay puts "--" in the empty column; after stripping amounts the
		// remaining text tells us which column had a value.
		// Strategy: check whether the amount appears before or after "--" in
		// the line, or fall back to description keywords.
		var amountStr, txType string

		if len(txAmounts) == 1 {
			amountStr = txAmounts[0]
			// Find position of the amount vs "--" separators
			amtIdx := strings.Index(line, txAmounts[0])
			// Count "--" occurrences before and after the amount
			beforeAmt := line[:amtIdx]
			afterAmt := line[amtIdx+len(txAmounts[0]):]
			dashBefore := strings.Count(beforeAmt, "--")
			dashAfter := strings.Count(afterAmt, "--")
			// OPay column order: Debit | Credit | Balance
			// If "--" appears before the amount → debit column was empty → credit
			// If "--" appears after the amount → credit column was empty → debit
			if dashBefore > dashAfter {
				txType = "credit"
			} else if dashAfter > dashBefore {
				txType = "debit"
			} else {
				// Tie-break on description keywords
				lower := strings.ToLower(line)
				if containsAny(lower, "transfer from", "credit", "salary", "reversal", "refund", "owealth withdrawal") {
					txType = "credit"
				} else {
					txType = "debit"
				}
			}
		} else {
			// Two amounts before balance: first = debit, second = credit
			d := parseAmount(txAmounts[0])
			cr := parseAmount(txAmounts[1])
			if cr > 0 && cr != d {
				amountStr = txAmounts[1]
				txType = "credit"
			} else {
				amountStr = txAmounts[0]
				txType = "debit"
			}
		}

		amount := parseAmount(amountStr)
		if amount <= 0 {
			continue
		}

		dateMatch := opayDateRe.FindString(line)
		merchant := extractOPayMerchant(line, amounts)
		rows = append(rows, []string{dateMatch, fmt.Sprintf("%.2f", amount), txType, merchant})
	}
	return rows
}

// extractOPayMerchant strips dates, amounts, "--", channel/ref noise and
// returns the Description field as the merchant.
func extractOPayMerchant(line string, amounts []string) string {
	m := opayDateRe.ReplaceAllString(line, "")
	for _, a := range amounts {
		m = strings.ReplaceAll(m, a, "")
	}
	// Remove "--" placeholders and common trailing tokens
	m = regexp.MustCompile(`--`).ReplaceAllString(m, "")
	m = regexp.MustCompile(`(?i)\b(Mobile|Web|USSD|ATM|POS)\b`).ReplaceAllString(m, "")
	// Remove long numeric reference numbers (10+ digits)
	m = regexp.MustCompile(`\b\d{10,}\b`).ReplaceAllString(m, "")
	m = strings.Join(strings.Fields(m), " ")
	if m == "" {
		return "Unknown"
	}
	return m
}

// ── Split-amount layout (GTBank, Access, Zenith) ──────────────────────────────
// Lines have a date, description, then separate debit/credit columns.
// DR/CR keyword may or may not be present.

func extractSplitAmountRows(lines []string) [][]string {
	var rows [][]string
	for _, line := range lines {
		if !dateRe.MatchString(line) {
			continue
		}
		amounts := amountRe.FindAllString(line, -1)
		if len(amounts) < 2 {
			continue
		}
		// Heuristic: if last amount is much larger it's a balance
		last := parseAmount(amounts[len(amounts)-1])
		prev := parseAmount(amounts[len(amounts)-2])
		if last < prev {
			// last is not a balance — fall through to tabular
			continue
		}

		txAmounts := amounts[:len(amounts)-1]
		var amountStr, txType string
		typeMatch := typeRe.FindString(line)
		if typeMatch != "" {
			amountStr = txAmounts[len(txAmounts)-1]
			if strings.EqualFold(typeMatch, "CR") || strings.EqualFold(typeMatch, "credit") {
				txType = "credit"
			} else {
				txType = "debit"
			}
		} else if len(txAmounts) >= 2 {
			d := parseAmount(txAmounts[len(txAmounts)-2])
			cr := parseAmount(txAmounts[len(txAmounts)-1])
			if cr > 0 && cr != d {
				amountStr = txAmounts[len(txAmounts)-1]
				txType = "credit"
			} else {
				amountStr = txAmounts[len(txAmounts)-2]
				txType = "debit"
			}
		} else {
			continue
		}

		amount := parseAmount(amountStr)
		if amount <= 0 {
			continue
		}
		dateMatch := dateRe.FindString(line)
		merchant := extractMerchant(line, dateMatch, amounts)
		rows = append(rows, []string{dateMatch, fmt.Sprintf("%.2f", amount), txType, merchant})
	}
	return rows
}

// ── Classic tabular layout ────────────────────────────────────────────────────
// Single amount + DR/CR keyword on the same line.

func extractTabularRows(lines []string) [][]string {
	var rows [][]string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !amountRe.MatchString(line) {
			continue
		}
		dateMatch := dateRe.FindString(line)
		if dateMatch == "" {
			continue
		}
		typeMatch := typeRe.FindString(line)
		if typeMatch == "" {
			continue
		}
		txType := "debit"
		if strings.EqualFold(typeMatch, "CR") || strings.EqualFold(typeMatch, "credit") {
			txType = "credit"
		}
		amountStr := strings.ReplaceAll(amountRe.FindString(line), ",", "")
		merchant := extractMerchant(line, dateMatch, amountRe.FindAllString(line, -1))
		rows = append(rows, []string{dateMatch, amountStr, txType, merchant})
	}
	return rows
}

// ── SMS-style layout ──────────────────────────────────────────────────────────

func extractSMSRows(lines []string) [][]string {
	var rows [][]string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		amount, txType, merchant, err := parseSMS(line)
		if err != nil {
			continue
		}
		rows = append(rows, []string{"", fmt.Sprintf("%.2f", amount), txType, merchant})
	}
	return rows
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func parseAmount(s string) float64 {
	v, _ := strconv.ParseFloat(strings.ReplaceAll(s, ",", ""), 64)
	return v
}

func extractMerchant(line, dateMatch string, amounts []string) string {
	m := line
	if dateMatch != "" {
		m = strings.ReplaceAll(m, dateMatch, "")
	}
	for _, a := range amounts {
		m = strings.ReplaceAll(m, a, "")
	}
	m = typeRe.ReplaceAllString(m, "")
	m = strings.Join(strings.Fields(m), " ")
	if m == "" {
		return "Unknown"
	}
	return m
}

func containsAny(s string, keywords ...string) bool {
	for _, kw := range keywords {
		if strings.Contains(s, kw) {
			return true
		}
	}
	return false
}

func splitLines(s string) []string {
	var lines []string
	for _, l := range strings.Split(s, "\n") {
		l = strings.TrimSpace(l)
		if l != "" {
			lines = append(lines, l)
		}
	}
	return lines
}

type pdfReaderAt struct{ b []byte }

func (r *pdfReaderAt) ReadAt(p []byte, off int64) (int, error) {
	if off >= int64(len(r.b)) {
		return 0, io.EOF
	}
	n := copy(p, r.b[off:])
	if n < len(p) {
		return n, io.EOF
	}
	return n, nil
}
