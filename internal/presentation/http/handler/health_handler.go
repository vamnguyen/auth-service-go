package handler

import (
	"context"
	"net/http"
	"time"

	"auth-service/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db *gorm.DB
}

func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

func (h *HealthHandler) Check(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	dbStatus := "up"
	sqlDB, err := h.db.DB()
	if err != nil || sqlDB.PingContext(ctx) != nil {
		dbStatus = "down"
	}

	status := "healthy"
	statusCode := http.StatusOK
	if dbStatus == "down" {
		status = "unhealthy"
		statusCode = http.StatusServiceUnavailable
	}

	response.Success(c, statusCode, gin.H{
		"status":   status,
		"database": dbStatus,
	})
}
