package notification

import (
	"context"

	"github.com/moninte/backend/internal/core"
)

// Service defines the notification business logic.
type Service interface {
	// Send persists an in-app notification AND fires an FCM push to all of
	// the user's registered device tokens.
	Send(ctx context.Context, userID string, notifType NotifType, title, body string, refID *string) (*Notification, error)
	// Create persists an in-app notification only (no push).
	Create(userID string, notifType NotifType, title, body string, refID *string) (*Notification, error)
	// List returns paginated notifications for a user.
	List(userID string, page, limit int) ([]*Notification, error)
	// UnreadCount returns the number of unread notifications.
	UnreadCount(userID string) (int, error)
	// MarkRead marks a single notification as read.
	MarkRead(id, userID string) error
	// MarkAllRead marks all of a user's notifications as read.
	MarkAllRead(userID string) error
	// Delete removes a notification.
	Delete(id, userID string) error
	// RegisterToken stores an FCM device token for a user.
	RegisterToken(userID, token, platform string) error
	// RemoveToken deletes an FCM device token (e.g. on logout).
	RemoveToken(userID, token string) error
}

type service struct {
	repo      Repository
	tokenRepo DeviceTokenRepository
	fcm       *FCMSender
}

// NewService returns a Service backed by the given repositories and FCM sender.
func NewService(repo Repository, tokenRepo DeviceTokenRepository, fcm *FCMSender) Service {
	return &service{repo: repo, tokenRepo: tokenRepo, fcm: fcm}
}

// Send persists the notification and fires FCM push to all user devices.
func (s *service) Send(ctx context.Context, userID string, notifType NotifType, title, body string, refID *string) (*Notification, error) {
	n, err := s.Create(userID, notifType, title, body, refID)
	if err != nil {
		return nil, err
	}
	// Fire push asynchronously so the caller is never blocked by FCM latency.
	go func() {
		tokens, err := s.tokenRepo.GetByUserID(userID)
		if err != nil || len(tokens) == 0 {
			return
		}
		data := map[string]string{
			"type":     string(notifType),
			"notif_id": n.ID,
		}
		if refID != nil {
			data["ref_id"] = *refID
		}
		s.fcm.SendToTokens(ctx, tokens, title, body, data)
	}()
	return n, nil
}

func (s *service) Create(userID string, notifType NotifType, title, body string, refID *string) (*Notification, error) {
	n := &Notification{
		UserScoped: core.UserScoped{UserID: userID},
		Type:       notifType,
		Title:      title,
		Body:       body,
		RefID:      refID,
	}
	return n, s.repo.Create(n)
}

func (s *service) List(userID string, page, limit int) ([]*Notification, error) {
	return s.repo.GetByUserID(userID, limit, (page-1)*limit)
}

func (s *service) UnreadCount(userID string) (int, error) {
	return s.repo.UnreadCount(userID)
}

func (s *service) MarkRead(id, userID string) error {
	return s.repo.MarkRead(id, userID)
}

func (s *service) MarkAllRead(userID string) error {
	return s.repo.MarkAllRead(userID)
}

func (s *service) Delete(id, userID string) error {
	return s.repo.Delete(id, userID)
}

func (s *service) RegisterToken(userID, token, platform string) error {
	return s.tokenRepo.Upsert(userID, token, platform)
}

func (s *service) RemoveToken(userID, token string) error {
	return s.tokenRepo.Delete(userID, token)
}
