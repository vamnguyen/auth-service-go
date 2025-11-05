package handler

import (
	"net/http"
	"strings"

	"auth-service/internal/application/dto"
	"auth-service/internal/application/usecase"
	domainErr "auth-service/internal/domain/error"
	"auth-service/internal/infrastructure/config"
	"auth-service/internal/infrastructure/logger"
	"auth-service/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthHandler struct {
	authUseCase *usecase.AuthUseCase
	config      *config.Config
	logger      *logger.Logger
}

func NewAuthHandler(authUseCase *usecase.AuthUseCase, config *config.Config, logger *logger.Logger) *AuthHandler {
	return &AuthHandler{
		authUseCase: authUseCase,
		config:      config,
		logger:      logger,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request payload")
		return
	}

	if err := h.authUseCase.Register(c.Request.Context(), req); err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, gin.H{
		"message": "User registered successfully",
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request payload")
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	result, err := h.authUseCase.Login(c.Request.Context(), req, ipAddress, userAgent)
	if err != nil {
		h.handleError(c, err)
		return
	}

	h.setRefreshCookie(c, result.RefreshToken)

	response.Success(c, http.StatusOK, gin.H{
		"access_token": result.AccessToken,
		"token_type":   result.TokenType,
		"expires_in":   result.ExpiresIn,
		"user":         result.User,
	})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	refreshToken, err := c.Cookie(h.config.Cookie.RefreshCookieName)
	if err != nil || refreshToken == "" {
		response.Unauthorized(c, "Missing refresh token")
		return
	}

	result, err := h.authUseCase.RefreshToken(c.Request.Context(), refreshToken)
	if err != nil {
		h.clearRefreshCookie(c)
		h.handleError(c, err)
		return
	}

	h.setRefreshCookie(c, result.RefreshToken)

	response.Success(c, http.StatusOK, gin.H{
		"access_token": result.AccessToken,
		"token_type":   result.TokenType,
		"expires_in":   result.ExpiresIn,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	userID := c.GetString("userID")
	refreshToken, _ := c.Cookie(h.config.Cookie.RefreshCookieName)
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	if err := h.authUseCase.Logout(c.Request.Context(), userID, refreshToken, ipAddress, userAgent); err != nil {
		h.logger.Error("Logout failed", zap.Error(err))
	}

	h.clearRefreshCookie(c)
	c.Status(http.StatusNoContent)
}

func (h *AuthHandler) LogoutAll(c *gin.Context) {
	userID := c.GetString("userID")
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	if err := h.authUseCase.LogoutAll(c.Request.Context(), userID, ipAddress, userAgent); err != nil {
		h.handleError(c, err)
		return
	}

	h.clearRefreshCookie(c)
	c.Status(http.StatusNoContent)
}

func (h *AuthHandler) GetMe(c *gin.Context) {
	userID := c.GetString("userID")

	user, err := h.authUseCase.GetMe(c.Request.Context(), userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, user)
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request payload")
		return
	}

	userID := c.GetString("userID")

	if err := h.authUseCase.ChangePassword(c.Request.Context(), userID, req); err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{
		"message": "Password changed successfully",
	})
}

func (h *AuthHandler) setRefreshCookie(c *gin.Context, refreshToken string) {
	sameSite := http.SameSiteLaxMode
	switch strings.ToLower(h.config.Cookie.SameSite) {
	case "strict":
		sameSite = http.SameSiteStrictMode
	case "none":
		sameSite = http.SameSiteNoneMode
	}

	c.SetSameSite(sameSite)
	c.SetCookie(
		h.config.Cookie.RefreshCookieName,
		refreshToken,
		int(h.config.JWT.RefreshTokenTTL.Seconds()),
		"/",
		h.config.Cookie.Domain,
		h.config.Cookie.Secure,
		true,
	)
}

func (h *AuthHandler) clearRefreshCookie(c *gin.Context) {
	c.SetCookie(
		h.config.Cookie.RefreshCookieName,
		"",
		-1,
		"/",
		h.config.Cookie.Domain,
		h.config.Cookie.Secure,
		true,
	)
}

func (h *AuthHandler) handleError(c *gin.Context, err error) {
	h.logger.Error("Request failed", zap.Error(err))

	switch err {
	case domainErr.ErrUserNotFound:
		response.NotFound(c, "User not found")
	case domainErr.ErrUserAlreadyExists:
		response.Conflict(c, "User already exists")
	case domainErr.ErrInvalidCredentials:
		response.Unauthorized(c, "Invalid credentials")
	case domainErr.ErrAccountLocked:
		response.Forbidden(c, "Account is locked. Please try again later")
	case domainErr.ErrInvalidToken, domainErr.ErrTokenExpired, domainErr.ErrTokenRevoked:
		response.Unauthorized(c, "Invalid or expired token")
	case domainErr.ErrWeakPassword:
		response.BadRequest(c, "Password is too weak. Must be at least 8 characters with uppercase, lowercase, number, and special character")
	case domainErr.ErrInvalidPassword:
		response.BadRequest(c, "Invalid password")
	default:
		response.InternalServerError(c, "An error occurred. Please try again")
	}
}
