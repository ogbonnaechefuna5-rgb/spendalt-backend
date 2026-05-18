package user

import (
	"errors"
	"regexp"
	"strings"

	"github.com/spendalt/backend/internal/lang"
)

var phoneRegex = regexp.MustCompile(`^\+?[0-9]{7,15}$`)

func validateUpdateProfile(firstName, lastName, phone string) error {
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
	return nil
}

func validateChangePassword(oldPassword, newPassword string) error {
	if strings.TrimSpace(oldPassword) == "" {
		return errors.New(lang.ErrCurrentPasswordRequired)
	}
	if strings.TrimSpace(newPassword) == "" {
		return errors.New(lang.ErrNewPasswordRequired)
	}
	if len(newPassword) < 8 {
		return errors.New(lang.ErrNewPasswordTooShort)
	}
	if oldPassword == newPassword {
		return errors.New(lang.ErrPasswordSameAsOld)
	}
	return nil
}
