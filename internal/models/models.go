package models

import "time"

type User struct {
	ID           int       `json:"id"`
	Phone        string    `json:"phone"`
	Email        string    `json:"email,omitempty"`
	FirstName    string    `json:"first_name,omitempty"`
	LastName     string    `json:"last_name,omitempty"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type RawTransaction struct {
	ID              int                    `json:"id"`
	UserID          int                    `json:"user_id"`
	Source          string                 `json:"source"`
	RawText         string                 `json:"raw_text,omitempty"`
	Amount          float64                `json:"amount"`
	TransactionType string                 `json:"transaction_type"`
	DetectedAt      *time.Time             `json:"detected_at,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	Processed       bool                   `json:"processed"`
	CreatedAt       time.Time              `json:"created_at"`
}

type Transaction struct {
	ID                int                    `json:"id"`
	UserID            int                    `json:"user_id"`
	RawTransactionID  *int                   `json:"raw_transaction_id,omitempty"`
	Amount            float64                `json:"amount"`
	Type              string                 `json:"type"`
	Merchant          string                 `json:"merchant"`
	Category          string                 `json:"category,omitempty"`
	Description       string                 `json:"description"`
	TransactionDate   time.Time              `json:"transaction_date"`
	BalanceAfter      *float64               `json:"balance_after,omitempty"`
	Fingerprint       string                 `json:"-"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
}

type Merchant struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	NormalizedName string    `json:"normalized_name"`
	Category       string    `json:"category,omitempty"`
	Aliases        []string  `json:"aliases,omitempty"`
	LogoURL        string    `json:"logo_url,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type Category struct {
	ID       int      `json:"id"`
	Name     string   `json:"name"`
	Icon     string   `json:"icon"`
	Color    string   `json:"color"`
	Keywords []string `json:"keywords"`
}

type CategoryBreakdown struct {
	Category string  `json:"category"`
	Total    float64 `json:"total"`
}

type Budget struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Category  string    `json:"category"`
	Amount    float64   `json:"amount"`
	Period    string    `json:"period"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	CreatedAt time.Time `json:"created_at"`
}

type SavingsGoal struct {
	ID            int        `json:"id"`
	UserID        int        `json:"user_id"`
	Name          string     `json:"name"`
	TargetAmount  float64    `json:"target_amount"`
	CurrentAmount float64    `json:"current_amount"`
	Deadline      *time.Time `json:"deadline,omitempty"`
	Status        string     `json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
}
