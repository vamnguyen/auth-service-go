package middleware

import (
	"net/http"

	"auth-service/internal/infrastructure/logger"
	"auth-service/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func RecoveryMiddleware(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
				)

				response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "An unexpected error occurred")
				c.Abort()
			}
		}()

		c.Next()
	}
}
