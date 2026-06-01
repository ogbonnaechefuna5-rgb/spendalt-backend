package notification

import (
	"time"

	"github.com/moninte/backend/internal/common"
)

// DeviceToken represents a user's FCM registration token for a specific device.
type DeviceToken struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Token     string    `json:"token"`
	Platform  string    `json:"platform"` // "android" | "ios" | "web"
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DeviceTokenRepository manages FCM token persistence.
type DeviceTokenRepository interface {
	Upsert(userID, token, platform string) error
	GetByUserID(userID string) ([]string, error)
	Delete(userID, token string) error
}

type deviceTokenRepository struct {
	db common.DB
}

// NewDeviceTokenRepository returns a Postgres-backed DeviceTokenRepository.
func NewDeviceTokenRepository(db common.DB) DeviceTokenRepository {
	return &deviceTokenRepository{db: db}
}

// Upsert inserts a token or updates its updated_at timestamp if it already exists.
func (r *deviceTokenRepository) Upsert(userID, token, platform string) error {
	_, err := r.db.Exec(
		`INSERT INTO device_tokens (user_id, token, platform)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (user_id, token) DO UPDATE SET updated_at = $4`,
		userID, token, platform, time.Now(),
	)
	return err
}

// GetByUserID returns all active FCM tokens for a user.
func (r *deviceTokenRepository) GetByUserID(userID string) ([]string, error) {
	rows, err := r.db.Query(
		`SELECT token FROM device_tokens WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tokens []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		tokens = append(tokens, t)
	}
	return tokens, nil
}

// Delete removes a specific token (e.g. on logout or token refresh).
func (r *deviceTokenRepository) Delete(userID, token string) error {
	_, err := r.db.Exec(
		`DELETE FROM device_tokens WHERE user_id = $1 AND token = $2`,
		userID, token,
	)
	return err
}
