package notification

import (
	"time"

	"github.com/moninte/backend/internal/common"
	"github.com/moninte/backend/internal/core"
)

// Repository defines persistence operations for notifications.
type Repository interface {
	Create(n *Notification) error
	GetByUserID(userID string, limit, offset int) ([]*Notification, error)
	UnreadCount(userID string) (int, error)
	MarkRead(id, userID string) error
	MarkAllRead(userID string) error
	Delete(id, userID string) error
}

type repository struct {
	db common.DB
}

// NewRepository returns a Postgres-backed Repository.
func NewRepository(db common.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(n *Notification) error {
	return r.db.QueryRow(
		`INSERT INTO notifications (user_id, type, title, body, ref_id)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, created_at`,
		n.UserID, n.Type, n.Title, n.Body, n.RefID,
	).Scan(&n.ID, &n.CreatedAt)
}

func (r *repository) GetByUserID(userID string, limit, offset int) ([]*Notification, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, type, title, body, read, read_at, ref_id, created_at
		 FROM notifications
		 WHERE user_id = $1
		 ORDER BY created_at DESC
		 LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	return core.ScanRows(rows, func(n *Notification) []interface{} {
		return []interface{}{
			&n.ID, &n.UserID, &n.Type, &n.Title, &n.Body,
			&n.Read, &n.ReadAt, &n.RefID, &n.CreatedAt,
		}
	})
}

func (r *repository) UnreadCount(userID string) (int, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND read = FALSE`,
		userID,
	).Scan(&count)
	return count, err
}

func (r *repository) MarkRead(id, userID string) error {
	now := time.Now()
	res, err := r.db.Exec(
		`UPDATE notifications SET read = TRUE, read_at = $1
		 WHERE id = $2 AND user_id = $3`,
		now, id, userID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return core.ErrNotFound
	}
	return nil
}

func (r *repository) MarkAllRead(userID string) error {
	now := time.Now()
	_, err := r.db.Exec(
		`UPDATE notifications SET read = TRUE, read_at = $1
		 WHERE user_id = $2 AND read = FALSE`,
		now, userID,
	)
	return err
}

func (r *repository) Delete(id, userID string) error {
	res, err := r.db.Exec(
		`DELETE FROM notifications WHERE id = $1 AND user_id = $2`,
		id, userID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return core.ErrNotFound
	}
	return nil
}
