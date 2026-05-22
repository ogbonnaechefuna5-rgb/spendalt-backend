package transaction

import (
	"time"

	"github.com/moninte/backend/internal/core"
)

type Transaction struct {
	core.UserScoped
	Amount           float64  `json:"amount"`
	Type             string   `json:"type"`
	Merchant         string   `json:"merchant"`
	Category         string   `json:"category,omitempty"`
	Description      string   `json:"description"`
	Source           string   `json:"source"`
	TransactionDate  time.Time `json:"transaction_date"`
	BalanceAfter     *float64  `json:"balance_after,omitempty"`
	RawTransactionID *string   `json:"-"`
	Fingerprint      string   `json:"-"`
}

type RawTransaction struct {
	ID              string     `json:"id"`
	UserID          string     `json:"user_id"`
	Source          string     `json:"source"`
	RawText         string     `json:"raw_text"`
	Amount          *float64   `json:"amount,omitempty"`
	TransactionType *string    `json:"transaction_type,omitempty"`
	DetectedAt      *time.Time `json:"detected_at,omitempty"`
	Processed       bool       `json:"processed"`
	CreatedAt       time.Time  `json:"created_at"`
}
