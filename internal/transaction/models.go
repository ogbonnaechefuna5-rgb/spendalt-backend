package transaction

import (
	"time"
	"github.com/spendalt/backend/internal/core"
)

type Transaction struct {
	core.UserScoped
	Amount          float64   `json:"amount"`
	Type            string    `json:"type"`
	Merchant        string    `json:"merchant"`
	Category        string    `json:"category,omitempty"`
	Description     string    `json:"description"`
	TransactionDate time.Time `json:"transaction_date"`
	Fingerprint     string    `json:"-"`
}