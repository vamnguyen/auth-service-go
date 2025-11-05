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

type UserModel struct {
	ID                  uuid.UUID  `gorm:"type:uuid;primaryKey"`
	Email               string     `gorm:"uniqueIndex;not null"`
	PasswordHash        string     `gorm:"column:password;not null"`
	Role                string     `gorm:"type:varchar(20);default:'user'"`
	IsVerified          bool       `gorm:"default:false"`
	IsLocked            bool       `gorm:"default:false"`
	FailedLoginAttempts int        `gorm:"default:0"`
	LockedUntil         *time.Time
	LastLoginAt         *time.Time
	LastLoginIP         string     `gorm:"type:varchar(45)"`
	CreatedAt           time.Time  `gorm:"autoCreateTime"`
	UpdatedAt           time.Time  `gorm:"autoUpdateTime"`
}

func (UserModel) TableName() string {
	return "users"
}

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	model := r.toModel(user)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return domainErr.ErrDatabaseOperation
	}
	return nil
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	var model UserModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainErr.ErrUserNotFound
		}
		return nil, domainErr.ErrDatabaseOperation
	}
	return r.toEntity(&model), nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var model UserModel
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainErr.ErrUserNotFound
		}
		return nil, domainErr.ErrDatabaseOperation
	}
	return r.toEntity(&model), nil
}

func (r *UserRepository) Update(ctx context.Context, user *entity.User) error {
	model := r.toModel(user)
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return domainErr.ErrDatabaseOperation
	}
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&UserModel{}, "id = ?", id).Error; err != nil {
		return domainErr.ErrDatabaseOperation
	}
	return nil
}

func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&UserModel{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return false, domainErr.ErrDatabaseOperation
	}
	return count > 0, nil
}

func (r *UserRepository) toModel(e *entity.User) *UserModel {
	return &UserModel{
		ID:                  e.ID,
		Email:               e.Email,
		PasswordHash:        e.PasswordHash,
		Role:                string(e.Role),
		IsVerified:          e.IsVerified,
		IsLocked:            e.IsLocked,
		FailedLoginAttempts: e.FailedLoginAttempts,
		LockedUntil:         e.LockedUntil,
		LastLoginAt:         e.LastLoginAt,
		LastLoginIP:         e.LastLoginIP,
		CreatedAt:           e.CreatedAt,
		UpdatedAt:           e.UpdatedAt,
	}
}

func (r *UserRepository) toEntity(m *UserModel) *entity.User {
	return &entity.User{
		ID:                  m.ID,
		Email:               m.Email,
		PasswordHash:        m.PasswordHash,
		Role:                entity.Role(m.Role),
		IsVerified:          m.IsVerified,
		IsLocked:            m.IsLocked,
		FailedLoginAttempts: m.FailedLoginAttempts,
		LockedUntil:         m.LockedUntil,
		LastLoginAt:         m.LastLoginAt,
		LastLoginIP:         m.LastLoginIP,
		CreatedAt:           m.CreatedAt,
		UpdatedAt:           m.UpdatedAt,
	}
}
