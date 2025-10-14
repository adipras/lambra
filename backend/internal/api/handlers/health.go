package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/yourusername/lambra/pkg/response"
)

type HealthHandler struct {
	db *sqlx.DB
}

func NewHealthHandler(db *sqlx.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

func (h *HealthHandler) HealthCheck(c *gin.Context) {
	// Check database connection
	if err := h.db.Ping(); err != nil {
		response.InternalError(c, "Database connection failed", err)
		return
	}

	response.Success(c, gin.H{
		"status": "healthy",
		"database": "connected",
	}, "Service is healthy")
}

func (h *HealthHandler) Readiness(c *gin.Context) {
	response.Success(c, gin.H{
		"status": "ready",
	}, "Service is ready")
}
