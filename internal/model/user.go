package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Email      string    `gorm:"uniqueIndex;not null" json:"email"`
	Password   string    `gorm:"not null" json:"-"`
	Role       string    `gorm:"type:varchar(20);default:'user'" json:"role"`
	IsVerified bool      `gorm:"default:false" json:"is_verified"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	RefreshTokens []RefreshToken `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

// BeforeCreate sets a default UUID (v7) before creating a new User record
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	u.ID = uuid.Must(uuid.NewV7())
	return
}
