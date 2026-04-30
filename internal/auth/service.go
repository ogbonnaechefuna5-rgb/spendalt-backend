package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/spendalt/backend/internal/common"
	"github.com/spendalt/backend/internal/user"
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
		return nil, errors.New("an account with this phone number already exists")
	}
	if email != "" {
		if _, err := s.userRepo.GetByEmail(email); err == nil {
			return nil, errors.New("an account with this email already exists")
		}
	}
	hashedPassword, err := common.HashPassword(password)
	if err != nil {
		return nil, errors.New("failed to process password")
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
		return nil, errors.New("failed to create account, please try again")
	}
	return u, nil
}

func (s *service) Login(identifier, password string, device, ip, deviceType, os, appVersion string) (*user.User, string, string, error) {
	if err := validateLogin(identifier, password); err != nil {
		return nil, "", "", err
	}
	u, err := s.userRepo.GetByPhone(identifier)
	if err != nil {
		u, err = s.userRepo.GetByEmail(identifier)
	}
	if err != nil {
		return nil, "", "", errors.New("incorrect phone/email or password")
	}
	if !common.CheckPassword(password, u.PasswordHash) {
		return nil, "", "", errors.New("incorrect phone number or password")
	}
	jti := uuid.NewString()
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
		return "", "", errors.New("invalid or expired refresh token")
	}
	if err := s.refreshRepo.Revoke(hash); err != nil {
		return "", "", err
	}
	newAccess, err := s.generateTokenWithJTI(rt.UserID, uuid.NewString())
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
