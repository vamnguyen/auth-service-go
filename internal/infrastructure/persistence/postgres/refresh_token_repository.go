package postgres

import (
	"context"
	"errors"
	"time"

	"auth-service/internal/domain/entity"
	domainErr "auth-service/internal/domain/error"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RefreshTokenModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	TokenHash string    `gorm:"column:token;uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"not null;index"`
	IsRevoked bool      `gorm:"column:revoked;default:false;index"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (RefreshTokenModel) TableName() string {
	return "refresh_tokens"
}

type RefreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, token *entity.RefreshToken) error {
	model := r.toModel(token)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return domainErr.ErrDatabaseOperation
	}
	return nil
}

func (r *RefreshTokenRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*entity.RefreshToken, error) {
	var model RefreshTokenModel
	if err := r.db.WithContext(ctx).Where("token = ?", tokenHash).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainErr.ErrInvalidToken
		}
		return nil, domainErr.ErrDatabaseOperation
	}
	return r.toEntity(&model), nil
}

func (r *RefreshTokenRepository) RevokeByTokenHash(ctx context.Context, tokenHash string) error {
	if err := r.db.WithContext(ctx).
		Model(&RefreshTokenModel{}).
		Where("token = ? AND revoked = FALSE", tokenHash).
		Update("revoked", true).Error; err != nil {
		return domainErr.ErrDatabaseOperation
	}
	return nil
}

func (r *RefreshTokenRepository) RevokeAllByUserID(ctx context.Context, userID uuid.UUID) error {
	if err := r.db.WithContext(ctx).
		Model(&RefreshTokenModel{}).
		Where("user_id = ? AND revoked = FALSE", userID).
		Update("revoked", true).Error; err != nil {
		return domainErr.ErrDatabaseOperation
	}
	return nil
}

func (r *RefreshTokenRepository) DeleteExpired(ctx context.Context) error {
	if err := r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&RefreshTokenModel{}).Error; err != nil {
		return domainErr.ErrDatabaseOperation
	}
	return nil
}

func (r *RefreshTokenRepository) toModel(e *entity.RefreshToken) *RefreshTokenModel {
	return &RefreshTokenModel{
		ID:        e.ID,
		UserID:    e.UserID,
		TokenHash: e.TokenHash,
		ExpiresAt: e.ExpiresAt,
		IsRevoked: e.IsRevoked,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}

func (r *RefreshTokenRepository) toEntity(m *RefreshTokenModel) *entity.RefreshToken {
	return &entity.RefreshToken{
		ID:        m.ID,
		UserID:    m.UserID,
		TokenHash: m.TokenHash,
		ExpiresAt: m.ExpiresAt,
		IsRevoked: m.IsRevoked,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}
