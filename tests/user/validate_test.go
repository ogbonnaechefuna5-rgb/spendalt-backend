package user_test

import (
	"errors"
	"testing"

	"github.com/moninte/backend/internal/lang"
	"github.com/moninte/backend/internal/user"
	"github.com/stretchr/testify/assert"
)

type validateProfileRepo struct{}

func (r *validateProfileRepo) Create(u *user.User) error                                          { return nil }
func (r *validateProfileRepo) GetByPhone(phone string) (*user.User, error)                        { return nil, errors.New("not found") }
func (r *validateProfileRepo) GetByEmail(email string) (*user.User, error)                        { return nil, errors.New("not found") }
func (r *validateProfileRepo) GetByID(id string) (*user.User, error)                              { return &user.User{}, nil }
func (r *validateProfileRepo) Update(u *user.User) error                                          { return nil }
func (r *validateProfileRepo) Delete(id string) error                                             { return nil }
func (r *validateProfileRepo) GetPreferences(userID string) (*user.UserPreferences, error)        { return &user.UserPreferences{}, nil }
func (r *validateProfileRepo) SavePreferences(userID string, sms, analytics, offers bool) error  { return nil }
func (r *validateProfileRepo) GetLinkedAccounts(userID string, limit, offset int) ([]*user.LinkedAccount, error) { return nil, nil }
func (r *validateProfileRepo) RemoveLinkedAccount(userID, accountID string) error                 { return nil }
func (r *validateProfileRepo) SyncLinkedAccount(userID, accountID string) error                   { return nil }
func (r *validateProfileRepo) GetSessions(userID string, limit, offset int) ([]*user.UserSession, error) { return nil, nil }
func (r *validateProfileRepo) RevokeAllSessions(userID string) error                              { return nil }
func (r *validateProfileRepo) RevokeSession(userID, sessionID string) error                       { return nil }
func (r *validateProfileRepo) CreateSession(s *user.UserSession) error                            { return nil }

type noopRefreshRevoker struct{}

func (r *noopRefreshRevoker) RevokeAllForUser(userID string) error { return nil }

func newValidateUserSvc() user.Service {
	return user.NewService(&validateProfileRepo{}, &noopRefreshRevoker{})
}

func TestUpdateProfile_MissingFirstName(t *testing.T) {
	err := newValidateUserSvc().UpdateProfile("u1", "", "", "Lovelace", "+2348012345678")
	assert.EqualError(t, err, lang.ErrFirstNameRequired)
}

func TestUpdateProfile_MissingLastName(t *testing.T) {
	err := newValidateUserSvc().UpdateProfile("u1", "Ada", "", "", "+2348012345678")
	assert.EqualError(t, err, lang.ErrLastNameRequired)
}

func TestUpdateProfile_MissingPhone(t *testing.T) {
	err := newValidateUserSvc().UpdateProfile("u1", "Ada", "", "Lovelace", "")
	assert.EqualError(t, err, lang.ErrPhoneRequired)
}

func TestUpdateProfile_InvalidPhone(t *testing.T) {
	err := newValidateUserSvc().UpdateProfile("u1", "Ada", "", "Lovelace", "notaphone")
	assert.EqualError(t, err, lang.ErrPhoneInvalid)
}

func TestChangePassword_MissingOld(t *testing.T) {
	err := newValidateUserSvc().ChangePassword("u1", "", "newpass1")
	assert.EqualError(t, err, lang.ErrCurrentPasswordRequired)
}

func TestChangePassword_MissingNew(t *testing.T) {
	err := newValidateUserSvc().ChangePassword("u1", "oldpass1", "")
	assert.EqualError(t, err, lang.ErrNewPasswordRequired)
}

func TestChangePassword_TooShort(t *testing.T) {
	err := newValidateUserSvc().ChangePassword("u1", "oldpass1", "short")
	assert.EqualError(t, err, lang.ErrNewPasswordTooShort)
}

func TestChangePassword_SameAsOld(t *testing.T) {
	err := newValidateUserSvc().ChangePassword("u1", "samepass", "samepass")
	assert.EqualError(t, err, lang.ErrPasswordSameAsOld)
}
