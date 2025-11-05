package controller

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"auth-service/config"
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
		ctx.JSON(http.StatusCreated, gin.H{"message": "user registered successfully"})
	}
}

func Login(authService *service.AuthService, cfg *config.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required"`
		}
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		access, refreshPlain, err := authService.Login(req.Email, req.Password)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		// Set refresh cookie
		setSameSite(ctx, cfg.CookieSameSite)
		ctx.SetCookie(
			cfg.RefreshCookieName,
			refreshPlain,
			int(cfg.RefreshTokenTTL.Seconds()),
			"/",
			cfg.CookieDomain,
			cfg.CookieSecure,
			true, // HttpOnly
		)

		ctx.JSON(http.StatusOK, gin.H{
			"access_token": access,
		})
	}
}

func Refresh(authService *service.AuthService, cfg *config.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		refreshPlain, err := ctx.Cookie(cfg.RefreshCookieName)
		if err != nil || refreshPlain == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "missing refresh token"})
			return
		}

		// Optional: if you want to bind refresh to a specific user, you can read current user from access (if present)
		accessUserID := "" // keep empty, service checks token record anyway

		newAccess, newRefreshPlain, err := authService.Refresh(refreshPlain, accessUserID)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		// Set rotated cookie
		setSameSite(ctx, cfg.CookieSameSite)
		ctx.SetCookie(
			cfg.RefreshCookieName,
			newRefreshPlain,
			int(cfg.RefreshTokenTTL.Seconds()),
			"/",
			cfg.CookieDomain,
			cfg.CookieSecure,
			true,
		)

		ctx.JSON(http.StatusOK, gin.H{
			"access_token": newAccess,
			"token_type":   "Bearer",
			"expires_in":   int(cfg.AccessTokenTTL.Seconds()),
		})
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

func Logout(authService *service.AuthService, cfg *config.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Logout phiên hiện tại theo refresh cookie
		refreshPlain, _ := ctx.Cookie(cfg.RefreshCookieName)
		if refreshPlain != "" {
			_ = authService.LogoutCurrent(refreshPlain)
		}
		// Xoá cookie
		clearCookie(ctx, cfg)
		ctx.Status(http.StatusNoContent)
	}
}

func LogoutAll(authService *service.AuthService, cfg *config.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID := ctx.GetString("userID")
		if userID == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		if err := authService.Logout(userID); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		clearCookie(ctx, cfg)
		ctx.Status(http.StatusNoContent)
	}
}

// Helpers =================================================

func setSameSite(c *gin.Context, mode string) {
	switch strings.ToLower(mode) {
	case "strict":
		c.SetSameSite(http.SameSiteStrictMode)
	case "none":
		// Lưu ý: SameSite=None yêu cầu CookieSecure=true
		c.SetSameSite(http.SameSiteNoneMode)
	default:
		c.SetSameSite(http.SameSiteLaxMode)
	}
}

func clearCookie(c *gin.Context, cfg *config.Config) {
	// MaxAge < 0 => delete cookie
	c.SetCookie(
		cfg.RefreshCookieName,
		"",
		-1,
		"/",
		cfg.CookieDomain,
		cfg.CookieSecure,
		true,
	)
}
