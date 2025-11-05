package middleware

import (
	"strings"

	"auth-service/internal/application/usecase"
	"auth-service/pkg/response"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(tokenService usecase.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "Missing authorization header")
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.Unauthorized(c, "Invalid authorization header format")
			c.Abort()
			return
		}

		token := parts[1]
		userID, err := tokenService.ValidateAccessToken(token)
		if err != nil {
			response.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		c.Set("userID", userID)
		c.Next()
	}
}
