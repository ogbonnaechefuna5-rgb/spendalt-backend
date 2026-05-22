package auth_test

import (
	"errors"
	"testing"
	"time"

	"github.com/moninte/backend/internal/auth"
	"github.com/moninte/backend/internal/lang"
	"github.com/moninte/backend/internal/user"
	"github.com/stretchr/testify/assert"
)

// minimal stubs — same as handler_test.go but scoped to this file
type validateUserRepo struct{}

func (r *validateUserRepo) Create(u *user.User) error                                          { u.ID = "uid-1"; return nil }
func (r *validateUserRepo) GetByPhone(phone string) (*user.User, error)                        { return nil, errors.New("not found") }
func (r *validateUserRepo) GetByEmail(email string) (*user.User, error)                        { return nil, errors.New("not found") }
func (r *validateUserRepo) GetByID(id string) (*user.User, error)                              { return &user.User{}, nil }
func (r *validateUserRepo) Update(u *user.User) error                                          { return nil }
func (r *validateUserRepo) UpdateAvatar(userID, avatarURL string) error                        { return nil }
func (r *validateUserRepo) Delete(id string) error                                             { return nil }
func (r *validateUserRepo) GetPreferences(userID string) (*user.UserPreferences, error)        { return &user.UserPreferences{}, nil }
func (r *validateUserRepo) SavePreferences(p *user.UserPreferences) error                 { return nil }
func (r *validateUserRepo) GetLinkedAccounts(userID string, limit, offset int) ([]*user.LinkedAccount, error) { return nil, nil }
func (r *validateUserRepo) RemoveLinkedAccount(userID, accountID string) error                 { return nil }
func (r *validateUserRepo) SyncLinkedAccount(userID, accountID string) error                   { return nil }
func (r *validateUserRepo) GetSessions(userID string, limit, offset int) ([]*user.UserSession, error) { return nil, nil }
func (r *validateUserRepo) RevokeAllSessions(userID string) error                              { return nil }
func (r *validateUserRepo) RevokeSession(userID, sessionID string) error                       { return nil }
func (r *validateUserRepo) CreateSession(s *user.UserSession) error                            { return nil }

type validateRefreshRepo struct{}

func (r *validateRefreshRepo) Store(userID, hash, device string, ttl time.Duration) error { return nil }
func (r *validateRefreshRepo) Get(hash string) (*auth.RefreshToken, error) {
	return &auth.RefreshToken{UserID: "uid-1", ExpiresAt: time.Now().Add(time.Hour)}, nil
}
func (r *validateRefreshRepo) Revoke(hash string) error             { return nil }
func (r *validateRefreshRepo) RevokeAllForUser(userID string) error { return nil }

func newValidateSvc() auth.Service {
	return auth.NewService(&validateUserRepo{}, "secret", auth.NewMemoryTokenStore(), &validateRefreshRepo{})
}

func TestRegister_MissingFirstName(t *testing.T) {
	_, err := newValidateSvc().Register("", "password1", "", "", "Lovelace", "+2348012345678")
	assert.EqualError(t, err, lang.ErrFirstNameRequired)
}

func TestRegister_MissingLastName(t *testing.T) {
	_, err := newValidateSvc().Register("", "password1", "Ada", "", "", "+2348012345678")
	assert.EqualError(t, err, lang.ErrLastNameRequired)
}

func TestRegister_MissingPhone(t *testing.T) {
	_, err := newValidateSvc().Register("", "password1", "Ada", "", "Lovelace", "")
	assert.EqualError(t, err, lang.ErrPhoneRequired)
}

func TestRegister_InvalidPhone(t *testing.T) {
	_, err := newValidateSvc().Register("", "password1", "Ada", "", "Lovelace", "abc")
	assert.EqualError(t, err, lang.ErrPhoneInvalid)
}

func TestRegister_MissingPassword(t *testing.T) {
	_, err := newValidateSvc().Register("", "", "Ada", "", "Lovelace", "+2348012345678")
	assert.EqualError(t, err, lang.ErrPasswordRequired)
}

func TestRegister_ShortPassword(t *testing.T) {
	_, err := newValidateSvc().Register("", "pass", "Ada", "", "Lovelace", "+2348012345678")
	assert.EqualError(t, err, lang.ErrPasswordTooShort)
}

func TestRegister_InvalidEmail(t *testing.T) {
	_, err := newValidateSvc().Register("notanemail", "password1", "Ada", "", "Lovelace", "+2348012345678")
	assert.EqualError(t, err, lang.ErrEmailInvalid)
}

func TestLogin_MissingIdentifierValidation(t *testing.T) {
	_, _, _, err := newValidateSvc().Login("", "password1", "", "", "", "", "")
	assert.EqualError(t, err, lang.ErrIdentifierRequired)
}

func TestLogin_MissingPasswordValidation(t *testing.T) {
	_, _, _, err := newValidateSvc().Login("+2348012345678", "", "", "", "", "", "")
	assert.EqualError(t, err, lang.ErrPasswordRequired)
}
