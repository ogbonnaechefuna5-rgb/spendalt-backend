package user

import (
	"time"
	"github.com/moninte/backend/internal/core"
)

type User struct {
	core.BaseModel
	Phone        string `json:"phone"`
	Email        string `json:"email,omitempty"`
	FirstName    string `json:"first_name,omitempty"`
	MiddleName   string `json:"middle_name,omitempty"`
	LastName     string `json:"last_name,omitempty"`
	AvatarURL    string `json:"avatar_url,omitempty"`
	PasswordHash string `json:"-"`
}

type UserPreferences struct {
	UserID              string `json:"user_id"`
	SMSDetection        bool   `json:"sms_detection"`
	Analytics           bool   `json:"analytics"`
	PartnerOffers       bool   `json:"partner_offers"`
	TransactionAlerts   bool   `json:"transaction_alerts"`
	BudgetWarnings      bool   `json:"budget_warnings"`
	AIInsights          bool   `json:"ai_insights"`
	WeeklyReport        bool   `json:"weekly_report"`
	SavingsReminders    bool   `json:"savings_reminders"`
	Promotions          bool   `json:"promotions"`
	HideBalances        bool   `json:"hide_balances"`
	CrashReports        bool   `json:"crash_reports"`
}

type LinkedAccount struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	BankName      string    `json:"bank_name"`
	AccountType   string    `json:"account_type"`
	AccountNumber string    `json:"account_number"`
	Balance       float64   `json:"balance"`
	Status        string    `json:"status"`
	LastSync      time.Time `json:"last_sync"`
	CreatedAt     time.Time `json:"created_at"`
}

type UserSession struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	TokenJTI    string    `json:"-"`
	Device      string    `json:"device"`
	DeviceType  string    `json:"device_type"`
	OS          string    `json:"os"`
	AppVersion  string    `json:"app_version"`
	IPAddress   string    `json:"ip_address"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}