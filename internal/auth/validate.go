package auth

import (
	"errors"
	"regexp"
	"strings"
)

var phoneRegex = regexp.MustCompile(`^\+?[0-9]{7,15}$`)

func validateRegister(email, password, firstName, lastName, phone string) error {
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
	if strings.TrimSpace(password) == "" {
		return errors.New("password is required")
	}
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	if email != "" && !strings.Contains(email, "@") {
		return errors.New("please enter a valid email address")
	}
	return nil
}

func validateLogin(identifier, password string) error {
	if strings.TrimSpace(identifier) == "" {
		return errors.New("phone number or email is required")
	}
	if strings.TrimSpace(password) == "" {
		return errors.New("password is required")
	}
	return nil
}
