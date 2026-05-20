package auth

import (
	"errors"
	"regexp"
	"strings"

	"github.com/moninte/backend/internal/lang"
)

var phoneRegex = regexp.MustCompile(`^\+?[0-9]{7,15}$`)

func validateRegister(email, password, firstName, lastName, phone string) error {
	if strings.TrimSpace(firstName) == "" {
		return errors.New(lang.ErrFirstNameRequired)
	}
	if strings.TrimSpace(lastName) == "" {
		return errors.New(lang.ErrLastNameRequired)
	}
	if strings.TrimSpace(phone) == "" {
		return errors.New(lang.ErrPhoneRequired)
	}
	if !phoneRegex.MatchString(phone) {
		return errors.New(lang.ErrPhoneInvalid)
	}
	if strings.TrimSpace(password) == "" {
		return errors.New(lang.ErrPasswordRequired)
	}
	if len(password) < 8 {
		return errors.New(lang.ErrPasswordTooShort)
	}
	if email != "" && !strings.Contains(email, "@") {
		return errors.New(lang.ErrEmailInvalid)
	}
	return nil
}

func validateLogin(identifier, password string) error {
	if strings.TrimSpace(identifier) == "" {
		return errors.New(lang.ErrIdentifierRequired)
	}
	if strings.TrimSpace(password) == "" {
		return errors.New(lang.ErrPasswordRequired)
	}
	return nil
}
