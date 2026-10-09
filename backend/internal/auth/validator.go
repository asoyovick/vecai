package auth

import (
	"errors"
	"net/mail"
	"strings"
	"unicode"
)

func validateRegistration(name, email, password string) (bool, error) {
	if strings.TrimSpace(name) == "" {
		return false, errors.New("Full name is required.")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return false, errors.New("Enter a valid email address.")
	}
	if err := ValidateNewPassword(password); err != nil {
		return false, err
	}
	return true, nil
}

// ValidateNewPassword enforces the password policy on changes/signups.
func ValidateNewPassword(password string) error {
	if len(password) < 8 {
		return errors.New("New password must be at least 8 characters.")
	}
	if len(password) > 200 {
		return errors.New("New password must be at most 200 characters.")
	}
	var hasLetter, hasDigit bool
	for _, r := range password {
		if unicode.IsLetter(r) {
			hasLetter = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return errors.New("New password must include at least one letter and one number.")
	}
	return nil
}
