package router

import (
	"auth-service/internal/application/usecase"
	"auth-service/internal/infrastructure/config"
	"auth-service/internal/infrastructure/logger"
	"auth-service/internal/presentation/http/handler"
	"auth-service/internal/presentation/middleware"

	"github.com/gin-gonic/gin"
)

type rateLimiter interface {
	Middleware() gin.HandlerFunc
}

type Router struct {
	authHandler   *handler.AuthHandler
	healthHandler *handler.HealthHandler
	tokenService  usecase.TokenService
	config        *config.Config
	logger        *logger.Logger
	rateLimiter   rateLimiter
}

func NewRouter(
	authHandler *handler.AuthHandler,
	healthHandler *handler.HealthHandler,
	tokenService usecase.TokenService,
	config *config.Config,
	logger *logger.Logger,
) *Router {
	return &Router{
		authHandler:   authHandler,
		healthHandler: healthHandler,
		tokenService:  tokenService,
		config:        config,
		logger:        logger,
		rateLimiter:   middleware.NewRateLimiter(config.Security.RateLimitPerMinute),
	}
}

func (r *Router) Setup() *gin.Engine {
	if r.config.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()

	engine.Use(middleware.RecoveryMiddleware(r.logger))
	engine.Use(middleware.LoggerMiddleware(r.logger))
	engine.Use(middleware.CORSMiddleware(r.config.Security.AllowedOrigins))

	engine.GET("/health", r.healthHandler.Check)

	api := engine.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", r.rateLimiter.Middleware(), r.authHandler.Register)
			auth.POST("/login", r.rateLimiter.Middleware(), r.authHandler.Login)
			auth.POST("/refresh", r.authHandler.RefreshToken)

			protected := auth.Group("")
			protected.Use(middleware.AuthMiddleware(r.tokenService))
			{
				protected.GET("/me", r.authHandler.GetMe)
				protected.POST("/logout", r.authHandler.Logout)
				protected.POST("/logout-all", r.authHandler.LogoutAll)
				protected.POST("/change-password", r.authHandler.ChangePassword)
			}
		}
	}

	return engine
}
