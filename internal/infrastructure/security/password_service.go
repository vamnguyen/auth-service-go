package security

import (
	"regexp"
	"unicode"

	domainErr "auth-service/internal/domain/error"
)

type PasswordService struct {
	minLength      int
	requireUpper   bool
	requireLower   bool
	requireNumber  bool
	requireSpecial bool
}

func NewPasswordService() *PasswordService {
	return &PasswordService{
		minLength:      8,
		requireUpper:   true,
		requireLower:   true,
		requireNumber:  true,
		requireSpecial: true,
	}
}

func (s *PasswordService) ValidateStrength(password string) error {
	if len(password) < s.minLength {
		return domainErr.ErrWeakPassword
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if s.requireUpper && !hasUpper {
		return domainErr.ErrWeakPassword
	}
	if s.requireLower && !hasLower {
		return domainErr.ErrWeakPassword
	}
	if s.requireNumber && !hasNumber {
		return domainErr.ErrWeakPassword
	}
	if s.requireSpecial && !hasSpecial {
		return domainErr.ErrWeakPassword
	}

	commonPasswords := []string{
		"password", "12345678", "qwerty", "abc123", "password123",
	}
	for _, common := range commonPasswords {
		if matched, _ := regexp.MatchString("(?i)"+common, password); matched {
			return domainErr.ErrWeakPassword
		}
	}

	return nil
}
