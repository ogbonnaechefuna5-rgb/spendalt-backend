package transaction_test

import (
	"testing"

	"github.com/moninte/backend/internal/transaction"
	"github.com/stretchr/testify/assert"
)

func TestCategorizer(t *testing.T) {
	c := transaction.NewRuleEngine()

	tests := []struct {
		merchant string
		want     string
	}{
		{"KFC Lagos", "Food & Dining"},
		{"UBER TRIP", "Transportation"},
		{"Shoprite Ikeja", "Shopping"},
		{"Netflix subscription", "Entertainment"},
		{"MTN Airtime", "Utilities"},
		{"Random Corp", "Other"},
		{"", "Other"},
		{"BOLT RIDE", "Transportation"},
		{"Dominos Pizza", "Food & Dining"},
	}

	for _, tt := range tests {
		t.Run(tt.merchant, func(t *testing.T) {
			assert.Equal(t, tt.want, c.Categorize(tt.merchant))
		})
	}
}
