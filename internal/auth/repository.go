package auth

import (
	"errors"
	"time"

	"github.com/spendalt/backend/internal/common"
)

type RefreshTokenRepository interface {
	Store(userID, hash, device string, ttl time.Duration) error
	Get(hash string) (*RefreshToken, error)
	Revoke(hash string) error
}

type refreshTokenRepo struct{ db common.DB }

func NewRefreshTokenRepository(db common.DB) RefreshTokenRepository {
	return &refreshTokenRepo{db: db}
}

func (r *refreshTokenRepo) Store(userID, hash, device string, ttl time.Duration) error {
	_, err := r.db.Exec(
		`INSERT INTO refresh_tokens (user_id, token_hash, device, expires_at) VALUES ($1, $2, $3, $4)`,
		userID, hash, device, time.Now().Add(ttl),
	)
	return err
}

func (r *refreshTokenRepo) Get(hash string) (*RefreshToken, error) {
	rt := &RefreshToken{}
	err := r.db.QueryRow(
		`SELECT user_id, device, expires_at, revoked FROM refresh_tokens WHERE token_hash = $1`, hash,
	).Scan(&rt.UserID, &rt.Device, &rt.ExpiresAt, &rt.Revoked)
	if err != nil {
		return nil, err
	}
	return rt, nil
}

func (r *refreshTokenRepo) Revoke(hash string) error {
	res, err := r.db.Exec(`UPDATE refresh_tokens SET revoked = TRUE WHERE token_hash = $1`, hash)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("token not found")
	}
	return nil
}
