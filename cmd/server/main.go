package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"auth-service/internal/application/usecase"
	"auth-service/internal/infrastructure/config"
	"auth-service/internal/infrastructure/logger"
	"auth-service/internal/infrastructure/persistence/postgres"
	"auth-service/internal/infrastructure/security"
	"auth-service/internal/presentation/http/handler"
	"auth-service/internal/presentation/http/router"

	"go.uber.org/zap"
	gormLogger "gorm.io/gorm/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	if err := cfg.Validate(); err != nil {
		panic(fmt.Sprintf("Config validation failed: %v", err))
	}

	log, err := logger.NewLogger(cfg.Environment)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer log.Sync()

	dbLogLevel := gormLogger.Info
	if cfg.Environment == "production" {
		dbLogLevel = gormLogger.Warn
	}

	db, err := postgres.NewDatabase(postgres.DatabaseConfig{
		URL:             cfg.Database.URL,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
		ConnMaxIdleTime: cfg.Database.ConnMaxIdleTime,
		LogLevel:        dbLogLevel,
	})
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}

	log.Info("Database connected successfully")

	if err := postgres.Migrate(db); err != nil {
		log.Fatal("Failed to migrate database", zap.Error(err))
	}
	log.Info("Database migrated successfully")

	userRepo := postgres.NewUserRepository(db)
	refreshTokenRepo := postgres.NewRefreshTokenRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)

	jwtService := security.NewJWTService(cfg.JWT.Secret, cfg.JWT.AccessTokenTTL)
	passwordService := security.NewPasswordService()

	authUseCase := usecase.NewAuthUseCase(
		userRepo,
		refreshTokenRepo,
		auditLogRepo,
		jwtService,
		passwordService,
		usecase.AuthConfig{
			AccessTokenTTL:      cfg.JWT.AccessTokenTTL,
			RefreshTokenTTL:     cfg.JWT.RefreshTokenTTL,
			MaxLoginAttempts:    cfg.Security.MaxLoginAttempts,
			AccountLockDuration: cfg.Security.AccountLockDuration,
		},
	)

	authHandler := handler.NewAuthHandler(authUseCase, cfg, log)
	healthHandler := handler.NewHealthHandler(db)

	r := router.NewRouter(authHandler, healthHandler, jwtService, cfg, log)
	engine := r.Setup()

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      engine,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		log.Info("Starting auth service", 
			zap.String("port", cfg.Server.Port),
			zap.String("environment", cfg.Environment),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown", zap.Error(err))
	}

	sqlDB, _ := db.DB()
	if err := sqlDB.Close(); err != nil {
		log.Error("Failed to close database connection", zap.Error(err))
	}

	log.Info("Server exited gracefully")
}
