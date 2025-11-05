package entity

import (
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	IsRevoked bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewRefreshToken(userID uuid.UUID, tokenHash string, ttl time.Duration) *RefreshToken {
	now := time.Now()
	return &RefreshToken{
		ID:        uuid.Must(uuid.NewV7()),
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: now.Add(ttl),
		IsRevoked: false,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (rt *RefreshToken) IsValid() bool {
	if rt.IsRevoked {
		return false
	}
	if time.Now().After(rt.ExpiresAt) {
		return false
	}
	return true
}

func (rt *RefreshToken) Revoke() {
	rt.IsRevoked = true
	rt.UpdatedAt = time.Now()
}
