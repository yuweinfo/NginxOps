package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthCheck 健康检查接口，与Java版本保持一致
// GET /api/health
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "UP",
	})
}
