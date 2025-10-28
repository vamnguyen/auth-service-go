package main

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"auth-service/config"
	"auth-service/internal/database"
	"auth-service/internal/repository"
	"auth-service/internal/router"
	"auth-service/internal/service"
)

func main() {
	cfg := config.LoadConfig()

	db, err := gorm.Open(postgres.Open(cfg.DBUrl), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	database.Migrate(db)

	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)

	r := router.SetupRouter(authService)
	log.Println("Auth service running on port " + cfg.Port)
	r.Run(":" + cfg.Port)
}
