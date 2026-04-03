package user

import (
	"errors"
	"github.com/spendalt/backend/internal/common"
)

type Service interface {
	GetProfile(userID int) (*User, error)
	UpdateProfile(userID int, firstName, lastName, phone string) error
	ChangePassword(userID int, oldPassword, newPassword string) error
	DeleteAccount(userID int) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetProfile(userID int) (*User, error) {
	return s.repo.GetByID(userID)
}

func (s *service) UpdateProfile(userID int, firstName, lastName, phone string) error {
	user, err := s.repo.GetByID(userID)
	if err != nil {
		return err
	}

	user.FirstName = firstName
	user.LastName = lastName
	user.Phone = phone

	return s.repo.Update(user)
}

func (s *service) ChangePassword(userID int, oldPassword, newPassword string) error {
	user, err := s.repo.GetByID(userID)
	if err != nil {
		return err
	}

	if !common.CheckPassword(oldPassword, user.PasswordHash) {
		return errors.New("invalid current password")
	}

	hashedPassword, err := common.HashPassword(newPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = hashedPassword
	return s.repo.Update(user)
}

func (s *service) DeleteAccount(userID int) error {
	return s.repo.Delete(userID)
}