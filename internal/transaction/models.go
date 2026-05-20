package transaction

import (
	"time"

	"github.com/moninte/backend/internal/core"
)

type Transaction struct {
	core.UserScoped
	Amount          float64   `json:"amount"`
	Type            string    `json:"type"`
	Merchant        string    `json:"merchant"`
	Category        string    `json:"category,omitempty"`
	Description     string    `json:"description"`
	Source          string    `json:"source"`
	TransactionDate time.Time `json:"transaction_date"`
	Fingerprint     string    `json:"-"`
}

type RawTransaction struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	Source     string    `json:"source"`
	RawText    string    `json:"raw_text"`
	Processed  bool      `json:"processed"`
	CreatedAt  time.Time `json:"created_at"`
}
