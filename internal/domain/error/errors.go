package error

import "errors"

var (
	// User errors
	ErrUserNotFound          = errors.New("user not found")
	ErrUserAlreadyExists     = errors.New("user already exists")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrAccountLocked         = errors.New("account is locked")
	ErrAccountNotVerified    = errors.New("account is not verified")
	ErrInvalidPassword       = errors.New("invalid password")
	ErrWeakPassword          = errors.New("password is too weak")

	// Token errors
	ErrInvalidToken          = errors.New("invalid token")
	ErrTokenExpired          = errors.New("token expired")
	ErrTokenRevoked          = errors.New("token revoked")
	ErrMissingToken          = errors.New("missing token")
	ErrInvalidTokenFormat    = errors.New("invalid token format")

	// Validation errors
	ErrInvalidEmail          = errors.New("invalid email format")
	ErrInvalidInput          = errors.New("invalid input")
	ErrMissingRequiredField  = errors.New("missing required field")

	// Infrastructure errors
	ErrDatabaseOperation     = errors.New("database operation failed")
	ErrCacheOperation        = errors.New("cache operation failed")
	ErrInternalServer        = errors.New("internal server error")

	// Rate limit errors
	ErrRateLimitExceeded     = errors.New("rate limit exceeded")
)
