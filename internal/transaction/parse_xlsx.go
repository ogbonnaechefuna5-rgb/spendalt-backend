package transaction

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

// parseXLSXBytes reads an xlsx file and returns rows in the same
// [date, amount, type, merchant] format expected by IngestCSV.
func parseXLSXBytes(data []byte) ([][]string, error) {
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("could not open xlsx: %w", err)
	}
	defer f.Close()

	// Use the first sheet
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("no sheets found in xlsx")
	}

	rawRows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("could not read sheet: %w", err)
	}
	if len(rawRows) == 0 {
		return nil, fmt.Errorf("empty sheet")
	}

	// Detect header row and column positions
	header := rawRows[0]
	cols := detectXLSXColumns(header)

	if cols == nil {
		// No recognisable header — treat every row as raw text and run through
		// the same PDF line-extraction logic
		var lines []string
		for _, row := range rawRows {
			lines = append(lines, strings.Join(row, "  "))
		}
		if rows := extractOPayRows(lines); len(rows) > 0 {
			return rows, nil
		}
		if rows := extractTabularRows(lines); len(rows) > 0 {
			return rows, nil
		}
		return nil, fmt.Errorf("could not detect transaction columns in xlsx")
	}

	var rows [][]string
	for _, row := range rawRows[1:] {
		if len(row) == 0 {
			continue
		}
		get := func(i int) string {
			if i < 0 || i >= len(row) {
				return ""
			}
			return strings.TrimSpace(row[i])
		}

		dateVal := get(cols.date)
		merchant := get(cols.merchant)
		if merchant == "" {
			merchant = "Unknown"
		}

		var amountStr, txType string

		if cols.debit >= 0 && cols.credit >= 0 {
			// Separate debit/credit columns
			d := parseAmount(get(cols.debit))
			cr := parseAmount(get(cols.credit))
			if cr > 0 {
				amountStr = fmt.Sprintf("%.2f", cr)
				txType = "credit"
			} else if d > 0 {
				amountStr = fmt.Sprintf("%.2f", d)
				txType = "debit"
			} else {
				continue
			}
		} else {
			// Single amount column + type column
			amountStr = get(cols.amount)
			raw := strings.ToLower(get(cols.txType))
			if strings.Contains(raw, "cr") || strings.Contains(raw, "credit") {
				txType = "credit"
			} else {
				txType = "debit"
			}
		}

		amount := parseAmount(amountStr)
		if amount <= 0 {
			continue
		}
		rows = append(rows, []string{dateVal, fmt.Sprintf("%.2f", amount), txType, merchant})
	}
	return rows, nil
}

type xlsxCols struct {
	date, amount, txType, merchant, debit, credit int
}

func detectXLSXColumns(header []string) *xlsxCols {
	cols := &xlsxCols{
		date: -1, amount: -1, txType: -1, merchant: -1, debit: -1, credit: -1,
	}
	for i, h := range header {
		h = strings.ToLower(strings.TrimSpace(h))
		switch {
		case contains(h, "date"):
			cols.date = i
		case contains(h, "debit", "dr amount", "withdrawal"):
			cols.debit = i
		case contains(h, "credit", "cr amount", "deposit"):
			cols.credit = i
		case contains(h, "amount"):
			cols.amount = i
		case contains(h, "type", "tran type", "transaction type"):
			cols.txType = i
		case contains(h, "description", "narration", "merchant", "details", "remark", "particulars"):
			cols.merchant = i
		}
	}
	// Need at least a date and either (debit+credit) or (amount)
	hasAmount := (cols.debit >= 0 && cols.credit >= 0) || cols.amount >= 0
	if cols.date < 0 || !hasAmount {
		return nil
	}
	return cols
}

func contains(s string, keywords ...string) bool {
	for _, kw := range keywords {
		if strings.Contains(s, kw) {
			return true
		}
	}
	return false
}
