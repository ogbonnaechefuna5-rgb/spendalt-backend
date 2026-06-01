package notification

import (
	"time"

	"github.com/moninte/backend/internal/core"
)

// NotifType classifies what triggered the notification.
type NotifType string

const (
	TypeTransaction NotifType = "transaction"
	TypeBudgetAlert NotifType = "budget_alert"
	TypeAIInsight   NotifType = "ai_insight"
	TypeSavings     NotifType = "savings"
	TypeSystem      NotifType = "system"
)

// Notification is a single in-app notification for a user.
type Notification struct {
	core.UserScoped
	Type   NotifType  `json:"type"`
	Title  string     `json:"title"`
	Body   string     `json:"body"`
	Read   bool       `json:"read"`
	ReadAt *time.Time `json:"read_at,omitempty"`
	// Optional deep-link reference (e.g. transaction ID, budget ID)
	RefID *string `json:"ref_id,omitempty"`
}
