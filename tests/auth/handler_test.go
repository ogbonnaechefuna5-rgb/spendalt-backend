package auth_test

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/moninte/backend/internal/auth"
	"github.com/moninte/backend/internal/common"
	"github.com/moninte/backend/internal/lang"
	"github.com/moninte/backend/internal/testutil"
	"github.com/moninte/backend/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── stub user repository ──────────────────────────────────────────────────────

type stubUserRepo struct {
	users map[string]*user.User // keyed by phone
}

func newStubUserRepo(users ...*user.User) *stubUserRepo {
	r := &stubUserRepo{users: make(map[string]*user.User)}
	for _, u := range users {
		r.users[u.Phone] = u
	}
	return r
}

func (r *stubUserRepo) Create(u *user.User) error {
	u.ID = testutil.TestUserID
	r.users[u.Phone] = u
	return nil
}
func (r *stubUserRepo) GetByPhone(phone string) (*user.User, error) {
	if u, ok := r.users[phone]; ok {
		return u, nil
	}
	return nil, errors.New("not found")
}
func (r *stubUserRepo) GetByEmail(email string) (*user.User, error) {
	for _, u := range r.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, errors.New("not found")
}
func (r *stubUserRepo) GetByID(id string) (*user.User, error)                             { return &user.User{}, nil }
func (r *stubUserRepo) Update(u *user.User) error                                          { return nil }
func (r *stubUserRepo) Delete(id string) error                                             { return nil }
func (r *stubUserRepo) GetPreferences(userID string) (*user.UserPreferences, error)        { return &user.UserPreferences{}, nil }
func (r *stubUserRepo) SavePreferences(userID string, sms, analytics, offers bool) error  { return nil }
func (r *stubUserRepo) GetLinkedAccounts(userID string, limit, offset int) ([]*user.LinkedAccount, error) {
	return nil, nil
}
func (r *stubUserRepo) RemoveLinkedAccount(userID, accountID string) error { return nil }
func (r *stubUserRepo) SyncLinkedAccount(userID, accountID string) error   { return nil }
func (r *stubUserRepo) GetSessions(userID string, limit, offset int) ([]*user.UserSession, error) {
	return nil, nil
}
func (r *stubUserRepo) RevokeAllSessions(userID string) error         { return nil }
func (r *stubUserRepo) RevokeSession(userID, sessionID string) error  { return nil }
func (r *stubUserRepo) CreateSession(s *user.UserSession) error       { return nil }

// ── stub refresh token repository ─────────────────────────────────────────────

type stubRefreshRepo struct{}

func (r *stubRefreshRepo) Store(userID, hash, device string, ttl time.Duration) error { return nil }
func (r *stubRefreshRepo) Get(hash string) (*auth.RefreshToken, error) {
	return &auth.RefreshToken{
		UserID:    testutil.TestUserID,
		ExpiresAt: time.Now().Add(time.Hour),
		Revoked:   false,
	}, nil
}
func (r *stubRefreshRepo) Revoke(hash string) error             { return nil }
func (r *stubRefreshRepo) RevokeAllForUser(userID string) error { return nil }

// ── app builder ───────────────────────────────────────────────────────────────

func newAuthApp(userRepo user.Repository) *fiber.App {
	svc := auth.NewService(userRepo, testutil.TestSecret, auth.NewMemoryTokenStore(), &stubRefreshRepo{})
	h := auth.NewHandler(svc)
	app := testutil.NewApp()
	app.Post("/auth/signup", h.Signup)
	app.Post("/auth/login", h.Login)
	app.Post("/auth/refresh", h.Refresh)
	app.Post("/auth/logout", h.Logout)
	return app
}

// ── signup ────────────────────────────────────────────────────────────────────

func TestSignup_Success(t *testing.T) {
	resp := testutil.Do(t, newAuthApp(newStubUserRepo()), http.MethodPost, "/auth/signup", auth.SignupRequest{
		FirstName: "Ada", LastName: "Lovelace", Phone: "+2348012345678", Password: "password1",
	}, "")
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var body map[string]any
	testutil.DecodeJSON(t, resp, &body)
	assert.Equal(t, "Account created successfully", body["message"])
}

func TestSignup_MissingFirstName(t *testing.T) {
	resp := testutil.Do(t, newAuthApp(newStubUserRepo()), http.MethodPost, "/auth/signup", auth.SignupRequest{
		LastName: "Lovelace", Phone: "+2348012345678", Password: "password1",
	}, "")
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	var body map[string]string
	testutil.DecodeJSON(t, resp, &body)
	assert.Equal(t, lang.ErrFirstNameRequired, body["error"])
}

func TestSignup_ShortPassword(t *testing.T) {
	resp := testutil.Do(t, newAuthApp(newStubUserRepo()), http.MethodPost, "/auth/signup", auth.SignupRequest{
		FirstName: "Ada", LastName: "Lovelace", Phone: "+2348012345678", Password: "short",
	}, "")
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	var body map[string]string
	testutil.DecodeJSON(t, resp, &body)
	assert.Equal(t, lang.ErrPasswordTooShort, body["error"])
}

func TestSignup_InvalidPhone(t *testing.T) {
	resp := testutil.Do(t, newAuthApp(newStubUserRepo()), http.MethodPost, "/auth/signup", auth.SignupRequest{
		FirstName: "Ada", LastName: "Lovelace", Phone: "notaphone", Password: "password1",
	}, "")
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	var body map[string]string
	testutil.DecodeJSON(t, resp, &body)
	assert.Equal(t, lang.ErrPhoneInvalid, body["error"])
}

func TestSignup_PhoneTaken(t *testing.T) {
	existing := &user.User{Phone: "+2348012345678"}
	resp := testutil.Do(t, newAuthApp(newStubUserRepo(existing)), http.MethodPost, "/auth/signup", auth.SignupRequest{
		FirstName: "Ada", LastName: "Lovelace", Phone: "+2348012345678", Password: "password1",
	}, "")
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	var body map[string]string
	testutil.DecodeJSON(t, resp, &body)
	assert.Equal(t, lang.ErrPhoneTaken, body["error"])
}

// ── login ─────────────────────────────────────────────────────────────────────

func TestLogin_Success(t *testing.T) {
	hash, err := common.HashPassword("password1")
	require.NoError(t, err)
	existing := &user.User{Phone: "+2348012345678", PasswordHash: hash}
	existing.ID = testutil.TestUserID

	resp := testutil.Do(t, newAuthApp(newStubUserRepo(existing)), http.MethodPost, "/auth/login", auth.LoginRequest{
		Identifier: "+2348012345678", Password: "password1",
	}, "")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var body map[string]any
	testutil.DecodeJSON(t, resp, &body)
	assert.NotEmpty(t, body["token"])
	assert.NotEmpty(t, body["refresh_token"])
	assert.NotNil(t, body["user"])
}

func TestLogin_WrongPassword(t *testing.T) {
	hash, err := common.HashPassword("password1")
	require.NoError(t, err)
	existing := &user.User{Phone: "+2348012345678", PasswordHash: hash}

	resp := testutil.Do(t, newAuthApp(newStubUserRepo(existing)), http.MethodPost, "/auth/login", auth.LoginRequest{
		Identifier: "+2348012345678", Password: "wrongpass",
	}, "")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	var body map[string]string
	testutil.DecodeJSON(t, resp, &body)
	assert.Equal(t, lang.ErrInvalidCredentials, body["error"])
}

func TestLogin_UserNotFound(t *testing.T) {
	resp := testutil.Do(t, newAuthApp(newStubUserRepo()), http.MethodPost, "/auth/login", auth.LoginRequest{
		Identifier: "+2348099999999", Password: "password1",
	}, "")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	var body map[string]string
	testutil.DecodeJSON(t, resp, &body)
	assert.Equal(t, lang.ErrInvalidCredentials, body["error"])
}

func TestLogin_MissingIdentifier(t *testing.T) {
	resp := testutil.Do(t, newAuthApp(newStubUserRepo()), http.MethodPost, "/auth/login", auth.LoginRequest{
		Password: "password1",
	}, "")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// ── refresh ───────────────────────────────────────────────────────────────────

func TestRefresh_Success(t *testing.T) {
	resp := testutil.Do(t, newAuthApp(newStubUserRepo()), http.MethodPost, "/auth/refresh", auth.RefreshRequest{
		RefreshToken: "some-token",
	}, "")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var body map[string]string
	testutil.DecodeJSON(t, resp, &body)
	assert.NotEmpty(t, body["token"])
	assert.NotEmpty(t, body["refresh_token"])
}

func TestRefresh_MissingToken(t *testing.T) {
	resp := testutil.Do(t, newAuthApp(newStubUserRepo()), http.MethodPost, "/auth/refresh", auth.RefreshRequest{}, "")
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// ── logout ────────────────────────────────────────────────────────────────────

func TestLogout_Success(t *testing.T) {
	resp := testutil.Do(t, newAuthApp(newStubUserRepo()), http.MethodPost, "/auth/logout", auth.LogoutRequest{
		RefreshToken: "some-token",
	}, "")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
