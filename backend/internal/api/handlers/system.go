package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SystemHandler struct {
	db *gorm.DB
}

func NewSystemHandler(db *gorm.DB) *SystemHandler {
	return &SystemHandler{db: db}
}

// Health check endpoint (public, minimal information for load balancers)
func (h *SystemHandler) Health(c *gin.Context) {
	// Test database connection
	sqlDB, err := h.db.DB()
	if err != nil {
		c.Status(http.StatusServiceUnavailable)
		return
	}

	if err := sqlDB.Ping(); err != nil {
		c.Status(http.StatusServiceUnavailable)
		return
	}

	c.Status(http.StatusOK)
}

// System status endpoint (protected, minimal information)
func (h *SystemHandler) Status(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "running",
	})
}

// DetailedStatus provides comprehensive system status (protected, admin only)
func (h *SystemHandler) DetailedStatus(c *gin.Context) {
	// Test database connection
	dbStatus := "connected"
	sqlDB, err := h.db.DB()
	if err != nil {
		dbStatus = "error"
	} else if err := sqlDB.Ping(); err != nil {
		dbStatus = "unreachable"
	}

	stats := gin.H{
		"service":  "RAG Dashboard API",
		"version":  "1.0.0",
		"status":   "running",
		"database": dbStatus,
	}

	// Add database connection stats if available
	if sqlDB != nil {
		dbStats := sqlDB.Stats()
		stats["database_connections"] = gin.H{
			"open":        dbStats.OpenConnections,
			"in_use":      dbStats.InUse,
			"idle":        dbStats.Idle,
			"max_open":    dbStats.MaxOpenConnections,
			"wait_count":  dbStats.WaitCount,
			"wait_duration": dbStats.WaitDuration.String(),
		}
	}

	c.JSON(http.StatusOK, stats)
}
