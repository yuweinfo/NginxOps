package handler

import (
	"nginxops/internal/service"
	"nginxops/pkg/response"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type StatsHandler struct {
	service *service.StatsService
}

func NewStatsHandler() *StatsHandler {
	return &StatsHandler{
		service: service.NewStatsService(),
	}
}

// GetDashboard 获取仪表盘数据
// GET /api/stats/dashboard
func (h *StatsHandler) GetDashboard(c *gin.Context) {
	data := h.service.GetDashboard()
	response.Success(c, data)
}

// QueryLogs 查询访问日志
// GET /api/stats/logs
func (h *StatsHandler) QueryLogs(c *gin.Context) {
	startStr := c.Query("start")
	endStr := c.Query("end")
	ip := c.Query("ip")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "50"))

	var start, end time.Time
	if startStr != "" {
		start, _ = time.Parse(time.RFC3339, startStr)
	}
	if endStr != "" {
		end, _ = time.Parse(time.RFC3339, endStr)
	}

	logs, total, err := h.service.QueryLogs(start, end, ip, page, size)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	// 转换为前端需要的格式
	records := make([]map[string]interface{}, 0)
	for _, log := range logs {
		records = append(records, map[string]interface{}{
			"id":       log.ID,
			"time":     log.TimeLocal.Format("2006-01-02 15:04:05"),
			"ip":       log.RemoteAddr,
			"method":   log.Method,
			"url":      log.Path,
			"status":   log.Status,
			"size":     log.BodyBytes,
			"userAgent": log.UserAgent,
			"referer":  log.Referer,
			"duration": int64(log.RT),
		})
	}

	response.Success(c, gin.H{
		"records": records,
		"total":   total,
		"page":    page,
		"size":    size,
	})
}
