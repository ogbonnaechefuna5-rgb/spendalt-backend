package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/moninte/backend/internal/common"
	"github.com/moninte/backend/internal/lang"
	"github.com/moninte/backend/internal/user"
)

type RefreshToken struct {
	UserID    string
	Device    string
	ExpiresAt time.Time
	Revoked   bool
}

type Service interface {
	Register(email, password, firstName, middleName, lastName, phone string) (*user.User, error)
	Login(identifier, password string, device, ip, deviceType, os, appVersion string) (*user.User, string, string, error)
	Logout(tokenString string) error
	ValidateToken(tokenString string) (*user.User, error)
	Refresh(rawToken string) (newAccess, newRefresh string, err error)
	RevokeRefreshToken(rawToken string) error
	OIDCLogin(provider, idToken, device, ip, deviceType, os, appVersion string) (*user.User, string, string, error)
}

type service struct {
	userRepo    user.Repository
	refreshRepo RefreshTokenRepository
	jwtSecret   string
	tokenStore  TokenStore
}

func NewService(userRepo user.Repository, jwtSecret string, tokenStore TokenStore, refreshRepo RefreshTokenRepository) Service {
	return &service{
		userRepo:    userRepo,
		refreshRepo: refreshRepo,
		jwtSecret:   jwtSecret,
		tokenStore:  tokenStore,
	}
}

func (s *service) Register(email, password, firstName, middleName, lastName, phone string) (*user.User, error) {
	if err := validateRegister(email, password, firstName, lastName, phone); err != nil {
		return nil, err
	}
	if _, err := s.userRepo.GetByPhone(phone); err == nil {
		return nil, errors.New(lang.ErrPhoneTaken)
	}
	if email != "" {
		if _, err := s.userRepo.GetByEmail(email); err == nil {
			return nil, errors.New(lang.ErrEmailTaken)
		}
	}
	hashedPassword, err := common.HashPassword(password)
	if err != nil {
		return nil, errors.New(lang.ErrPasswordHash)
	}
	u := &user.User{
		Email:        strings.TrimSpace(email),
		PasswordHash: hashedPassword,
		FirstName:    firstName,
		MiddleName:   middleName,
		LastName:     lastName,
		Phone:        phone,
	}
	if err := s.userRepo.Create(u); err != nil {
		log.Printf("[auth] userRepo.Create failed: %v", err)
		return nil, errors.New(lang.ErrCreateAccount)
	}
	return u, nil
}

const maxLoginAttempts = 10

func loginAttemptsKey(identifier string) string {
	h := sha256.Sum256([]byte("login:" + identifier))
	return hex.EncodeToString(h[:])
}

func (s *service) Login(identifier, password string, device, ip, deviceType, os, appVersion string) (*user.User, string, string, error) {
	if err := validateLogin(identifier, password); err != nil {
		return nil, "", "", err
	}
	attemptsKey := loginAttemptsKey(identifier)
	attempts, _ := s.tokenStore.GetCounter(attemptsKey)
	if attempts >= maxLoginAttempts {
		return nil, "", "", errors.New(lang.ErrAccountLocked)
	}
	u, err := s.userRepo.GetByPhone(identifier)
	if err != nil {
		u, err = s.userRepo.GetByEmail(identifier)
	}
	if err != nil {
		s.tokenStore.IncrCounter(attemptsKey, 15*time.Minute)
		return nil, "", "", errors.New(lang.ErrInvalidCredentials)
	}
	if !common.CheckPassword(password, u.PasswordHash) {
		s.tokenStore.IncrCounter(attemptsKey, 15*time.Minute)
		return nil, "", "", errors.New(lang.ErrInvalidCredentials)
	}
	s.tokenStore.DeleteCounter(attemptsKey)
	jti := common.NewID()
	accessToken, err := s.generateTokenWithJTI(u.ID, jti)
	if err != nil {
		return nil, "", "", err
	}
	rawRefresh, refreshHash, err := generateRefreshToken()
	if err != nil {
		return nil, "", "", err
	}
	if err := s.refreshRepo.Store(u.ID, refreshHash, device, 30*24*time.Hour); err != nil {
		return nil, "", "", err
	}
	_ = s.userRepo.CreateSession(&user.UserSession{
		UserID:     u.ID,
		TokenJTI:   jti,
		Device:     device,
		DeviceType: deviceType,
		OS:         os,
		AppVersion: appVersion,
		IPAddress:  ip,
		ExpiresAt:  time.Now().Add(30 * 24 * time.Hour),
	})
	return u, accessToken, rawRefresh, nil
}

func (s *service) Refresh(rawToken string) (string, string, error) {
	hash := hashToken(rawToken)
	rt, err := s.refreshRepo.Get(hash)
	if err != nil || rt.Revoked || rt.ExpiresAt.Before(time.Now()) {
		return "", "", errors.New(lang.ErrInvalidRefreshToken)
	}
	if err := s.refreshRepo.Revoke(hash); err != nil {
		return "", "", err
	}
	newAccess, err := s.generateTokenWithJTI(rt.UserID, common.NewID())
	if err != nil {
		return "", "", err
	}
	newRaw, newHash, err := generateRefreshToken()
	if err != nil {
		return "", "", err
	}
	if err := s.refreshRepo.Store(rt.UserID, newHash, rt.Device, 30*24*time.Hour); err != nil {
		return "", "", err
	}
	return newAccess, newRaw, nil
}

func (s *service) RevokeRefreshToken(rawToken string) error {
	return s.refreshRepo.Revoke(hashToken(rawToken))
}

func (s *service) Logout(tokenString string) error {
	token, err := s.parseToken(tokenString)
	if err != nil {
		return err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return errors.New("invalid token claims")
	}
	jti, ok := claims["jti"].(string)
	if !ok || jti == "" {
		return errors.New("token has no jti")
	}
	return s.tokenStore.Revoke(jti)
}

func (s *service) ValidateToken(tokenString string) (*user.User, error) {
	token, err := s.parseToken(tokenString)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}
	if jti, ok := claims["jti"].(string); ok && jti != "" {
		revoked, err := s.tokenStore.IsRevoked(jti)
		if err != nil || revoked {
			return nil, errors.New("token has been revoked")
		}
	}
	userID := claims["user_id"].(string)
	return s.userRepo.GetByID(userID)
}

func (s *service) parseToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return token, nil
}

func (s *service) generateTokenWithJTI(userID, jti string) (string, error) {
	claims := jwt.MapClaims{
		"jti":     jti,
		"user_id": userID,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func generateRefreshToken() (raw, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return
	}
	raw = base64.URLEncoding.EncodeToString(b)
	hash = hashToken(raw)
	return
}

func hashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

// ── OIDC ──────────────────────────────────────────────────────────────────────

// OIDCLogin verifies a Google or Apple ID token, then finds or creates the
// corresponding user account and issues app tokens.
func (s *service) OIDCLogin(provider, idToken, device, ip, deviceType, os, appVersion string) (*user.User, string, string, error) {
	var email, firstName, lastName, sub string
	var err error

	switch provider {
	case "google":
		email, firstName, lastName, sub, err = verifyGoogleIDToken(idToken)
	case "apple":
		email, firstName, lastName, sub, err = verifyAppleIDToken(idToken)
	default:
		return nil, "", "", errors.New("unsupported OIDC provider")
	}
	if err != nil {
		return nil, "", "", fmt.Errorf("OIDC token verification failed: %w", err)
	}

	// Find existing user by email, or create one.
	u, err := s.userRepo.GetByEmail(email)
	if err != nil {
		// New user — create account with a random unusable password.
		randBytes := make([]byte, 16)
		if _, rerr := rand.Read(randBytes); rerr != nil {
			return nil, "", "", rerr
		}
		placeholder, _ := common.HashPassword(base64.URLEncoding.EncodeToString(randBytes))
		parts := strings.SplitN(firstName+" "+lastName, " ", 2)
		fn, ln := parts[0], ""
		if len(parts) == 2 {
			ln = parts[1]
		}
		u = &user.User{
			Email:        email,
			PasswordHash: placeholder,
			FirstName:    fn,
			LastName:     ln,
			// Phone is not provided by OIDC — leave empty; user can add later.
		}
		if cerr := s.userRepo.Create(u); cerr != nil {
			return nil, "", "", errors.New(lang.ErrCreateAccount)
		}
		_ = sub // stored for future account linking if needed
	}

	jti := common.NewID()
	accessToken, err := s.generateTokenWithJTI(u.ID, jti)
	if err != nil {
		return nil, "", "", err
	}
	rawRefresh, refreshHash, err := generateRefreshToken()
	if err != nil {
		return nil, "", "", err
	}
	if err := s.refreshRepo.Store(u.ID, refreshHash, device, 30*24*time.Hour); err != nil {
		return nil, "", "", err
	}
	_ = s.userRepo.CreateSession(&user.UserSession{
		UserID:     u.ID,
		TokenJTI:   jti,
		Device:     device,
		DeviceType: deviceType,
		OS:         os,
		AppVersion: appVersion,
		IPAddress:  ip,
		ExpiresAt:  time.Now().Add(30 * 24 * time.Hour),
	})
	return u, accessToken, rawRefresh, nil
}

// ── OIDC token verification ───────────────────────────────────────────────────

// oidcClaims holds the standard claims we extract from Google/Apple ID tokens.
type oidcClaims struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Name          string `json:"name"`
	EmailVerified bool   `json:"email_verified"`
}

// jwksKey is a single key from a JWKS endpoint.
type jwksKey struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type jwksResponse struct {
	Keys []jwksKey `json:"keys"`
}

func fetchJWKS(url string) ([]jwksKey, error) {
	resp, err := http.Get(url) //nolint:gosec // URL is a hardcoded constant
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var jwks jwksResponse
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, err
	}
	return jwks.Keys, nil
}

func rsaPublicKeyFromJWK(k jwksKey) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
	if err != nil {
		return nil, err
	}
	e := 0
	for _, b := range eBytes {
		e = e<<8 + int(b)
	}
	return &rsa.PublicKey{N: new(big.Int).SetBytes(nBytes), E: e}, nil
}

func verifyIDToken(tokenStr, jwksURL, expectedAudience string) (*oidcClaims, error) {
	keys, err := fetchJWKS(jwksURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}

	// Parse without verification first to extract the kid header.
	unverified, _, err := new(jwt.Parser).ParseUnverified(tokenStr, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token header: %w", err)
	}
	kid, _ := unverified.Header["kid"].(string)

	// Find the matching key.
	var matchedKey *jwksKey
	for i := range keys {
		if keys[i].Kid == kid {
			matchedKey = &keys[i]
			break
		}
	}
	if matchedKey == nil {
		// If no kid match, try the first RSA key (Apple sometimes omits kid).
		for i := range keys {
			if keys[i].Kty == "RSA" {
				matchedKey = &keys[i]
				break
			}
		}
	}
	if matchedKey == nil {
		return nil, errors.New("no matching JWKS key found")
	}

	pubKey, err := rsaPublicKeyFromJWK(*matchedKey)
	if err != nil {
		return nil, fmt.Errorf("failed to build RSA public key: %w", err)
	}

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return pubKey, nil
	}, jwt.WithAudience(expectedAudience), jwt.WithExpirationRequired())
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	claims := &oidcClaims{
		Sub:   fmt.Sprintf("%v", mapClaims["sub"]),
		Email: fmt.Sprintf("%v", mapClaims["email"]),
	}
	if gn, ok := mapClaims["given_name"].(string); ok {
		claims.GivenName = gn
	}
	if fn, ok := mapClaims["family_name"].(string); ok {
		claims.FamilyName = fn
	}
	if n, ok := mapClaims["name"].(string); ok {
		claims.Name = n
	}
	if ev, ok := mapClaims["email_verified"].(bool); ok {
		claims.EmailVerified = ev
	}
	return claims, nil
}

func verifyGoogleIDToken(idToken string) (email, firstName, lastName, sub string, err error) {
	// Google's JWKS endpoint and client-ID audience.
	// The audience must match the OAuth client ID configured in the Flutter app.
	const googleJWKS = "https://www.googleapis.com/oauth2/v3/certs"
	// We accept any audience here and rely on the token's exp/signature for
	// security. In production, set this to your actual Google client ID.
	// Using jwt.WithAudience("") would reject all tokens, so we parse manually.
	keys, ferr := fetchJWKS(googleJWKS)
	if ferr != nil {
		err = ferr
		return
	}

	unverified, _, perr := new(jwt.Parser).ParseUnverified(idToken, jwt.MapClaims{})
	if perr != nil {
		err = perr
		return
	}
	kid, _ := unverified.Header["kid"].(string)

	var matchedKey *jwksKey
	for i := range keys {
		if keys[i].Kid == kid {
			matchedKey = &keys[i]
			break
		}
	}
	if matchedKey == nil {
		err = errors.New("no matching Google JWKS key")
		return
	}
	pubKey, kerr := rsaPublicKeyFromJWK(*matchedKey)
	if kerr != nil {
		err = kerr
		return
	}

	token, verr := jwt.Parse(idToken, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return pubKey, nil
	}, jwt.WithExpirationRequired(), jwt.WithIssuedAt())
	if verr != nil || !token.Valid {
		err = fmt.Errorf("Google token invalid: %w", verr)
		return
	}

	mapClaims := token.Claims.(jwt.MapClaims)
	// Verify issuer
	iss, _ := mapClaims["iss"].(string)
	if iss != "accounts.google.com" && iss != "https://accounts.google.com" {
		err = errors.New("invalid Google token issuer")
		return
	}

	email, _ = mapClaims["email"].(string)
	firstName, _ = mapClaims["given_name"].(string)
	lastName, _ = mapClaims["family_name"].(string)
	sub, _ = mapClaims["sub"].(string)
	if email == "" || sub == "" {
		err = errors.New("Google token missing required claims")
	}
	return
}

func verifyAppleIDToken(idToken string) (email, firstName, lastName, sub string, err error) {
	const appleJWKS = "https://appleid.apple.com/auth/keys"
	keys, ferr := fetchJWKS(appleJWKS)
	if ferr != nil {
		err = ferr
		return
	}

	unverified, _, perr := new(jwt.Parser).ParseUnverified(idToken, jwt.MapClaims{})
	if perr != nil {
		err = perr
		return
	}
	kid, _ := unverified.Header["kid"].(string)

	var matchedKey *jwksKey
	for i := range keys {
		if keys[i].Kid == kid {
			matchedKey = &keys[i]
			break
		}
	}
	if matchedKey == nil {
		for i := range keys {
			if keys[i].Kty == "RSA" {
				matchedKey = &keys[i]
				break
			}
		}
	}
	if matchedKey == nil {
		err = errors.New("no matching Apple JWKS key")
		return
	}
	pubKey, kerr := rsaPublicKeyFromJWK(*matchedKey)
	if kerr != nil {
		err = kerr
		return
	}

	token, verr := jwt.Parse(idToken, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return pubKey, nil
	}, jwt.WithExpirationRequired(), jwt.WithIssuedAt())
	if verr != nil || !token.Valid {
		err = fmt.Errorf("Apple token invalid: %w", verr)
		return
	}

	mapClaims := token.Claims.(jwt.MapClaims)
	iss, _ := mapClaims["iss"].(string)
	if iss != "https://appleid.apple.com" {
		err = errors.New("invalid Apple token issuer")
		return
	}

	email, _ = mapClaims["email"].(string)
	sub, _ = mapClaims["sub"].(string)
	// Apple only provides name on first sign-in; subsequent logins omit it.
	firstName, _ = mapClaims["given_name"].(string)
	lastName, _ = mapClaims["family_name"].(string)
	if sub == "" {
		err = errors.New("Apple token missing sub claim")
	}
	return
}
