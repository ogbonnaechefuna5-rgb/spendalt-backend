package main

import (
	"fmt"
	"os"
	"strings"
	"strconv"
	"github.com/moninte/backend/internal/transaction"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: testparse <file.pdf>")
		return
	}
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Println("read error:", err)
		return
	}
	rows, err := transaction.TestPDFParse(data)
	fmt.Printf("parsed=%d err=%v\n\n", len(rows), err)

	skipped, ok := 0, 0
	for _, row := range rows {
		if len(row) < 4 { skipped++; fmt.Printf("SKIP short row: %v\n", row); continue }
		amtStr := strings.ReplaceAll(strings.TrimSpace(row[1]), ",", "")
		amt, aerr := strconv.ParseFloat(amtStr, 64)
		if aerr != nil || amt <= 0 { skipped++; fmt.Printf("SKIP bad amount: %q\n", row[1]); continue }
		txType := strings.ToLower(strings.TrimSpace(row[2]))
		if txType != "debit" && txType != "credit" { skipped++; fmt.Printf("SKIP bad type: %q\n", row[2]); continue }
		merchant := strings.TrimSpace(row[3])
		if merchant == "" { skipped++; fmt.Printf("SKIP empty merchant\n"); continue }
		ok++
	}
	fmt.Printf("\nwould_import=%d would_skip=%d\n", ok, skipped)
}
