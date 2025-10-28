package database

import (
	"log"

	"auth-service/internal/model"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) {
	if err := db.AutoMigrate(
		&model.User{},
		&model.RefreshToken{},
	); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}
	log.Println("✅ Database migrated successfully")
}
