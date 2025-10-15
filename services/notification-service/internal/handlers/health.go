package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// HealthCheck handles health check requests
func HealthCheck(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Basic health check - service is running
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "notification-service",
			"version":   "1.0.0",
			"timestamp": gin.H{},
		})
	}
}
