package user

import (
	"errors"
	"regexp"
	"strings"
)

var phoneRegex = regexp.MustCompile(`^\+?[0-9]{7,15}$`)

func validateUpdateProfile(firstName, lastName, phone string) error {
	if strings.TrimSpace(firstName) == "" {
		return errors.New("first name is required")
	}
	if strings.TrimSpace(lastName) == "" {
		return errors.New("last name is required")
	}
	if strings.TrimSpace(phone) == "" {
		return errors.New("phone number is required")
	}
	if !phoneRegex.MatchString(phone) {
		return errors.New("please enter a valid phone number")
	}
	return nil
}

func validateChangePassword(oldPassword, newPassword string) error {
	if strings.TrimSpace(oldPassword) == "" {
		return errors.New("current password is required")
	}
	if strings.TrimSpace(newPassword) == "" {
		return errors.New("new password is required")
	}
	if len(newPassword) < 8 {
		return errors.New("new password must be at least 8 characters")
	}
	if oldPassword == newPassword {
		return errors.New("new password must be different from your current password")
	}
	return nil
}
