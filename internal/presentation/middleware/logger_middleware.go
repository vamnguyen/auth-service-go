package middleware

import (
	"time"

	"auth-service/internal/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func LoggerMiddleware(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		userAgent := c.Request.UserAgent()

		fields := []zap.Field{
			zap.String("method", method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("client_ip", clientIP),
			zap.String("user_agent", userAgent),
		}

		if len(c.Errors) > 0 {
			log.Error("Request completed with errors", append(fields, zap.String("errors", c.Errors.String()))...)
		} else if statusCode >= 500 {
			log.Error("Request failed", fields...)
		} else if statusCode >= 400 {
			log.Warn("Request client error", fields...)
		} else {
			log.Info("Request completed", fields...)
		}
	}
}
