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

func (r *RefreshTokenRepository) Create(rt *model.RefreshToken) error {
	return r.DB.Create(rt).Error
}

func (r *RefreshTokenRepository) FindByTokenHash(tokenHash string) (*model.RefreshToken, error) {
	var rt model.RefreshToken
	if err := r.DB.Where("token = ?", tokenHash).First(&rt).Error; err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *RefreshTokenRepository) RevokeByTokenHash(tokenHash string) error {
	return r.DB.Model(&model.RefreshToken{}).
		Where("token = ? AND revoked = FALSE", tokenHash).
		Update("revoked", true).Error
}

func (r *RefreshTokenRepository) RevokeAllByUser(userID uuid.UUID) error {
	return r.DB.Model(&model.RefreshToken{}).
		Where("user_id = ? AND revoked = FALSE", userID).
		Update("revoked", true).Error
}
