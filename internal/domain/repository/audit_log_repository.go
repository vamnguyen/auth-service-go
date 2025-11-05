package repository

import (
	"context"

	"auth-service/internal/domain/entity"
)

type AuditLogRepository interface {
	Create(ctx context.Context, log *entity.AuditLog) error
}
