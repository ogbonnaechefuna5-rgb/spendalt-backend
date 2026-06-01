package notification

import "github.com/moninte/backend/internal/core"

// ListResponse is the shape of GET /notifications.
type ListResponse struct {
	Notifications []*Notification `json:"notifications"`
	UnreadCount   int             `json:"unread_count"`
	core.PageMeta
}
