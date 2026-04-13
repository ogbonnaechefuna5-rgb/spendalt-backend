package user

import (
	"errors"
	"strings"

	"github.com/spendalt/backend/internal/common"
)

type Service interface {
	GetProfile(userID string) (*User, error)
	UpdateProfile(userID string, firstName, middleName, lastName, phone string) error
	ChangePassword(userID string, oldPassword, newPassword string) error
	DeleteAccount(userID string) error
	GetPreferences(userID string) (*UserPreferences, error)
	SavePreferences(userID string, sms, analytics, offers bool) error
	GetLinkedAccounts(userID string, limit, offset int) ([]*LinkedAccount, error)
	RemoveLinkedAccount(userID, accountID string) error
	SyncLinkedAccount(userID, accountID string) error
	GetSessions(userID string, limit, offset int) ([]*UserSession, error)
	RevokeAllSessions(userID string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetProfile(userID string) (*User, error) {
	return s.repo.GetByID(userID)
}

func (s *service) UpdateProfile(userID string, firstName, middleName, lastName, phone string) error {
	firstName = strings.TrimSpace(firstName)
	middleName = strings.TrimSpace(middleName)
	lastName = strings.TrimSpace(lastName)
	phone = strings.TrimSpace(phone)
	if err := validateUpdateProfile(firstName, lastName, phone); err != nil {
		return err
	}
	user, err := s.repo.GetByID(userID)
	if err != nil {
		return err
	}
	user.FirstName = firstName
	user.MiddleName = middleName
	user.LastName = lastName
	user.Phone = phone
	return s.repo.Update(user)
}

func (s *service) ChangePassword(userID string, oldPassword, newPassword string) error {
	if err := validateChangePassword(oldPassword, newPassword); err != nil {
		return err
	}
	user, err := s.repo.GetByID(userID)
	if err != nil {
		return err
	}

	if !common.CheckPassword(oldPassword, user.PasswordHash) {
		return errors.New("current password is incorrect")
	}

	hashedPassword, err := common.HashPassword(newPassword)
	if err != nil {
		return err
	}
	user.PasswordHash = hashedPassword
	return s.repo.Update(user)
}

func (s *service) DeleteAccount(userID string) error {
	return s.repo.Delete(userID)
}

func (s *service) GetPreferences(userID string) (*UserPreferences, error) {
	return s.repo.GetPreferences(userID)
}

func (s *service) SavePreferences(userID string, sms, analytics, offers bool) error {
	return s.repo.SavePreferences(userID, sms, analytics, offers)
}

func (s *service) GetLinkedAccounts(userID string, limit, offset int) ([]*LinkedAccount, error) {
	return s.repo.GetLinkedAccounts(userID, limit, offset)
}

func (s *service) RemoveLinkedAccount(userID, accountID string) error {
	return s.repo.RemoveLinkedAccount(userID, accountID)
}

func (s *service) SyncLinkedAccount(userID, accountID string) error {
	return s.repo.SyncLinkedAccount(userID, accountID)
}

func (s *service) GetSessions(userID string, limit, offset int) ([]*UserSession, error) {
	return s.repo.GetSessions(userID, limit, offset)
}

func (s *service) RevokeAllSessions(userID string) error {
	return s.repo.RevokeAllSessions(userID)
}
