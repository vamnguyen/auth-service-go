package entity

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID                  uuid.UUID
	Email               string
	PasswordHash        string
	Role                Role
	IsVerified          bool
	IsLocked            bool
	FailedLoginAttempts int
	LockedUntil         *time.Time
	LastLoginAt         *time.Time
	LastLoginIP         string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

func NewUser(email, password string) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &User{
		ID:           uuid.Must(uuid.NewV7()),
		Email:        email,
		PasswordHash: string(hashedPassword),
		Role:         RoleUser,
		IsVerified:   false,
		IsLocked:     false,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func (u *User) VerifyPassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
}

func (u *User) IsAccountLocked() bool {
	if !u.IsLocked {
		return false
	}
	if u.LockedUntil != nil && time.Now().After(*u.LockedUntil) {
		return false
	}
	return true
}

func (u *User) IncrementFailedLoginAttempts(maxAttempts int, lockDuration time.Duration) {
	u.FailedLoginAttempts++
	if u.FailedLoginAttempts >= maxAttempts {
		u.IsLocked = true
		lockedUntil := time.Now().Add(lockDuration)
		u.LockedUntil = &lockedUntil
	}
}

func (u *User) ResetFailedLoginAttempts() {
	u.FailedLoginAttempts = 0
	u.IsLocked = false
	u.LockedUntil = nil
}

func (u *User) UpdateLastLogin(ipAddress string) {
	now := time.Now()
	u.LastLoginAt = &now
	u.LastLoginIP = ipAddress
}

func (u *User) ChangePassword(newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hashedPassword)
	u.UpdatedAt = time.Now()
	return nil
}

func (u *User) Verify() {
	u.IsVerified = true
	u.UpdatedAt = time.Now()
}
