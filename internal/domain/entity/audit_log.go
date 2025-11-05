package entity

import (
	"time"

	"github.com/google/uuid"
)

type AuditLog struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Action    AuditAction
	IPAddress string
	UserAgent string
	Metadata  map[string]interface{}
	CreatedAt time.Time
}

type AuditAction string

const (
	AuditActionLogin             AuditAction = "login"
	AuditActionLoginFailed       AuditAction = "login_failed"
	AuditActionLogout            AuditAction = "logout"
	AuditActionRegister          AuditAction = "register"
	AuditActionPasswordChange    AuditAction = "password_change"
	AuditActionPasswordReset     AuditAction = "password_reset"
	AuditActionEmailVerification AuditAction = "email_verification"
	AuditActionTokenRefresh      AuditAction = "token_refresh"
	AuditActionAccountLocked     AuditAction = "account_locked"
)

func NewAuditLog(userID uuid.UUID, action AuditAction, ipAddress, userAgent string) *AuditLog {
	return &AuditLog{
		ID:        uuid.Must(uuid.NewV7()),
		UserID:    userID,
		Action:    action,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Metadata:  make(map[string]interface{}),
		CreatedAt: time.Now(),
	}
}

func (a *AuditLog) AddMetadata(key string, value interface{}) {
	if a.Metadata == nil {
		a.Metadata = make(map[string]interface{})
	}
	a.Metadata[key] = value
}
