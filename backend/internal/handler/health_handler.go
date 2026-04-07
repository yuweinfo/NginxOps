package handler

import (
	"nginxops/internal/service"
	"nginxops/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	checker *service.HealthChecker
}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{
		checker: service.GetHealthChecker(),
	}
}

// CheckUpstream 立即检查指定 upstream 的健康状态
// POST /api/upstreams/:id/health-check
func (h *HealthHandler) CheckUpstream(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	results, err := h.checker.CheckUpstreamNow(uint(id))
	if err != nil {
		response.Error(c, 200, err.Error())
		return
	}

	response.Success(c, results)
}

// StartHealthChecker 启动健康检查服务
func StartHealthChecker(intervalSeconds int) {
	service.GetHealthChecker().Start(intervalSeconds)
}

// StopHealthChecker 停止健康检查服务
func StopHealthChecker() {
	service.GetHealthChecker().Stop()
}
