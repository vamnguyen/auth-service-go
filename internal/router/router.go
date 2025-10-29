package router

import (
	"github.com/gin-gonic/gin"

	"auth-service/internal/controller"
	"auth-service/internal/middleware"
	"auth-service/internal/service"
)

func SetupRouter(authService *service.AuthService) *gin.Engine {
	r := gin.Default()

	r.GET("/health", controller.CheckHealth(authService))
	r.POST("/register", controller.Register(authService))
	r.POST("/login", controller.Login(authService))

	// Protected routes
	auth := r.Group("/")
	auth.Use(middleware.AuthMiddleware(authService.JWTSecret))

	auth.GET("/me", controller.GetMe(authService))
	auth.POST("/logout", controller.Logout(authService))

	return r
}
