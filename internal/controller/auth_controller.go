package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"auth-service/internal/service"
)

func CheckHealth(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	}
}

func Register(authService *service.AuthService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required,min=6"`
		}

		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		if err := authService.Register(req.Email, req.Password); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"message": "user registered successfully"})
	}
}

func Login(authService *service.AuthService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required"`
		}

		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		token, err := authService.Login(req.Email, req.Password)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"token": token})
	}
}

func GetMe(authService *service.AuthService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID := ctx.GetString("userID")
		user, err := authService.GetMe(userID)
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{
			"id":          user.ID,
			"email":       user.Email,
			"role":        user.Role,
			"is_verified": user.IsVerified,
			"created_at":  user.CreatedAt,
			"updated_at":  user.UpdatedAt,
		})
	}
}

func Logout(authService *service.AuthService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID := ctx.GetString("userID")
		if err := authService.Logout(userID); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "logged out"})
	}
}
