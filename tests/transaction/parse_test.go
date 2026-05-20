package transaction_test

import (
	"testing"
	"time"

	"github.com/moninte/backend/internal/transaction"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSMS(t *testing.T) {
	tests := []struct {
		name         string
		sms          string
		wantAmount   float64
		wantType     string
		wantMerchant string
		wantErr      bool
	}{
		{
			name:         "debit with merchant",
			sms:          "Your account has been debited NGN 12,500.00 at Shoprite on 01/01/2025",
			wantAmount:   12500.0,
			wantType:     "debit",
			wantMerchant: "Shoprite",
		},
		{
			name:         "credit transaction",
			sms:          "Credit alert: NGN 50000 received into your account",
			wantAmount:   50000.0,
			wantType:     "credit",
			wantMerchant: "Unknown",
		},
		{
			name:         "amount with no comma",
			sms:          "Debit: NGN 2500 at Bolt",
			wantAmount:   2500.0,
			wantType:     "debit",
			wantMerchant: "Bolt",
		},
		{
			name:    "no NGN amount",
			sms:     "Your OTP is 123456",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			amount, txType, merchant, err := transaction.ParseSMS(tt.sms)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantAmount, amount)
			assert.Equal(t, tt.wantType, txType)
			assert.Equal(t, tt.wantMerchant, merchant)
		})
	}
}

func TestGenerateFingerprint(t *testing.T) {
	date := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	fp1 := transaction.GenerateFingerprint("user1", 100.0, "Shoprite", date)
	fp2 := transaction.GenerateFingerprint("user1", 100.0, "Shoprite", date)
	fp3 := transaction.GenerateFingerprint("user2", 100.0, "Shoprite", date)
	fp4 := transaction.GenerateFingerprint("user1", 200.0, "Shoprite", date)

	assert.Equal(t, fp1, fp2, "same inputs must produce same fingerprint")
	assert.NotEqual(t, fp1, fp3, "different user must differ")
	assert.NotEqual(t, fp1, fp4, "different amount must differ")
	assert.Len(t, fp1, 64, "SHA-256 hex is 64 chars")
}
