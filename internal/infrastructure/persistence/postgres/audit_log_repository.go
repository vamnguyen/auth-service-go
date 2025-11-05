package postgres

import (
	"context"
	"encoding/json"
	"time"

	"auth-service/internal/domain/entity"
	domainErr "auth-service/internal/domain/error"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuditLogModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	Action    string    `gorm:"type:varchar(50);not null;index"`
	IPAddress string    `gorm:"type:varchar(45)"`
	UserAgent string    `gorm:"type:text"`
	Metadata  string    `gorm:"type:jsonb"`
	CreatedAt time.Time `gorm:"autoCreateTime;index"`
}

func (AuditLogModel) TableName() string {
	return "audit_logs"
}

type AuditLogRepository struct {
	db *gorm.DB
}

func NewAuditLogRepository(db *gorm.DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

func (r *AuditLogRepository) Create(ctx context.Context, log *entity.AuditLog) error {
	model := r.toModel(log)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return domainErr.ErrDatabaseOperation
	}
	return nil
}

func (r *AuditLogRepository) toModel(e *entity.AuditLog) *AuditLogModel {
	metadataJSON, _ := json.Marshal(e.Metadata)
	return &AuditLogModel{
		ID:        e.ID,
		UserID:    e.UserID,
		Action:    string(e.Action),
		IPAddress: e.IPAddress,
		UserAgent: e.UserAgent,
		Metadata:  string(metadataJSON),
		CreatedAt: e.CreatedAt,
	}
}
