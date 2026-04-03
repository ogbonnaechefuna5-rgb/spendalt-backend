package auth

import (
	"errors"
	"time"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spendalt/backend/internal/user"
	"github.com/spendalt/backend/internal/common"
)

type Service interface {
	Register(email, password, firstName, lastName, phone string) (*user.User, error)
	Login(email, password string) (*user.User, string, error)
	ValidateToken(token string) (*user.User, error)
}

type service struct {
	userRepo  user.Repository
	jwtSecret string
}

func NewService(userRepo user.Repository, jwtSecret string) Service {
	return &service{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

func (s *service) Register(email, password, firstName, lastName, phone string) (*user.User, error) {
	if _, err := s.userRepo.GetByEmail(email); err == nil {
		return nil, errors.New("user already exists")
	}

	hashedPassword, err := common.HashPassword(password)
	if err != nil {
		return nil, err
	}

	u := &user.User{
		Email:        email,
		PasswordHash: hashedPassword,
		FirstName:    firstName,
		LastName:     lastName,
		Phone:        phone,
	}

	if err := s.userRepo.Create(u); err != nil {
		return nil, err
	}

	return u, nil
}

func (s *service) Login(email, password string) (*user.User, string, error) {
	u, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, "", errors.New("invalid credentials")
	}

	if !common.CheckPassword(password, u.PasswordHash) {
		return nil, "", errors.New("invalid credentials")
	}

	token, err := s.generateToken(u.ID)
	if err != nil {
		return nil, "", err
	}

	return u, token, nil
}

func (s *service) ValidateToken(tokenString string) (*user.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	userID := int(claims["user_id"].(float64))
	return s.userRepo.GetByID(userID)
}

func (s *service) generateToken(userID int) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24 * 30).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}