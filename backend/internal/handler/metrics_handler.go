package handler

import (
	"fmt"
	"nginxops/internal/service"
	"nginxops/pkg/response"
	"time"

	"github.com/gin-gonic/gin"
)

const maxMetricsRangeDays = 90

type MetricsHandler struct {
	service *service.MetricsService
}

func NewMetricsHandler() *MetricsHandler {
	return &MetricsHandler{
		service: service.NewMetricsService(),
	}
}

// parseTimeRange 解析并校验时间范围参数，返回 start, end 或错误
func parseMetricsTimeRange(c *gin.Context) (time.Time, time.Time, bool) {
	startStr := c.Query("start")
	endStr := c.Query("end")

	start, _ := time.Parse(time.RFC3339, startStr)
	end, _ := time.Parse(time.RFC3339, endStr)

	// 默认时间范围：最近 24 小时
	if start.IsZero() {
		end = time.Now()
		start = end.Add(-24 * time.Hour)
	}
	if end.IsZero() {
		end = time.Now()
	}

	// 校验时间范围
	if end.Before(start) {
		response.BadRequest(c, "结束时间不能早于开始时间")
		return time.Time{}, time.Time{}, true
	}

	if end.Sub(start) > time.Duration(maxMetricsRangeDays)*24*time.Hour {
		response.BadRequest(c, fmt.Sprintf("查询时间范围不能超过 %d 天", maxMetricsRangeDays))
		return time.Time{}, time.Time{}, true
	}

	return start, end, false
}

func (h *MetricsHandler) GetOverview(c *gin.Context) {
	start, end, aborted := parseMetricsTimeRange(c)
	if aborted {
		return
	}

	data := h.service.GetOverview(start, end)
	response.Success(c, data)
}

func (h *MetricsHandler) GetTrafficTrend(c *gin.Context) {
	start, end, aborted := parseMetricsTimeRange(c)
	if aborted {
		return
	}
	granularity := c.DefaultQuery("granularity", "5m")

	data := h.service.GetTrafficTrend(start, end, granularity)
	response.Success(c, data)
}

func (h *MetricsHandler) GetResponseTrend(c *gin.Context) {
	start, end, aborted := parseMetricsTimeRange(c)
	if aborted {
		return
	}
	granularity := c.DefaultQuery("granularity", "5m")

	data := h.service.GetResponseTrend(start, end, granularity)
	response.Success(c, data)
}

func (h *MetricsHandler) GetSlowRequestTrend(c *gin.Context) {
	start, end, aborted := parseMetricsTimeRange(c)
	if aborted {
		return
	}
	granularity := c.DefaultQuery("granularity", "5m")

	data := h.service.GetSlowRequestTrend(start, end, granularity)
	response.Success(c, data)
}

func (h *MetricsHandler) GetMethodDistribution(c *gin.Context) {
	start, end, aborted := parseMetricsTimeRange(c)
	if aborted {
		return
	}

	data := h.service.GetMethodDistribution(start, end)
	response.Success(c, data)
}

func (h *MetricsHandler) GetStatusDistribution(c *gin.Context) {
	start, end, aborted := parseMetricsTimeRange(c)
	if aborted {
		return
	}

	data := h.service.GetStatusDistribution(start, end)
	response.Success(c, data)
}

func (h *MetricsHandler) GetErrorRateTrend(c *gin.Context) {
	start, end, aborted := parseMetricsTimeRange(c)
	if aborted {
		return
	}
	granularity := c.DefaultQuery("granularity", "5m")

	data := h.service.GetErrorRateTrend(start, end, granularity)
	response.Success(c, data)
}

func (h *MetricsHandler) GetErrorPaths(c *gin.Context) {
	start, end, aborted := parseMetricsTimeRange(c)
	if aborted {
		return
	}

	data := h.service.GetErrorPaths(start, end, 20)
	response.Success(c, data)
}

func (h *MetricsHandler) GetClientAnalysis(c *gin.Context) {
	start, end, aborted := parseMetricsTimeRange(c)
	if aborted {
		return
	}

	data := h.service.GetClientAnalysis(start, end)
	response.Success(c, data)
}
