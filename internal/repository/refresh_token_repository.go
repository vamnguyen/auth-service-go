package repository

import (
	"auth-service/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RefreshTokenRepository struct {
	DB *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{DB: db}
}

func (r *RefreshTokenRepository) RevokeAllByUser(userID uuid.UUID) error {
	return r.DB.Model(&model.RefreshToken{}).
		Where("user_id = ? AND revoked = FALSE", userID).
		Update("revoked", true).Error
}
