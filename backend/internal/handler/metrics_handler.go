package handler

import (
	"nginxops/internal/service"
	"nginxops/pkg/response"
	"time"

	"github.com/gin-gonic/gin"
)

type MetricsHandler struct {
	service *service.MetricsService
}

func NewMetricsHandler() *MetricsHandler {
	return &MetricsHandler{
		service: service.NewMetricsService(),
	}
}

func (h *MetricsHandler) GetOverview(c *gin.Context) {
	startStr := c.Query("start")
	endStr := c.Query("end")

	start, _ := time.Parse(time.RFC3339, startStr)
	end, _ := time.Parse(time.RFC3339, endStr)

	if start.IsZero() {
		end = time.Now()
		start = end.Add(-24 * time.Hour)
	}

	data := h.service.GetOverview(start, end)
	response.Success(c, data)
}

func (h *MetricsHandler) GetTrafficTrend(c *gin.Context) {
	startStr := c.Query("start")
	endStr := c.Query("end")
	granularity := c.DefaultQuery("granularity", "5m")

	start, _ := time.Parse(time.RFC3339, startStr)
	end, _ := time.Parse(time.RFC3339, endStr)

	if start.IsZero() {
		end = time.Now()
		start = end.Add(-24 * time.Hour)
	}

	data := h.service.GetTrafficTrend(start, end, granularity)
	response.Success(c, data)
}

func (h *MetricsHandler) GetResponseTrend(c *gin.Context) {
	startStr := c.Query("start")
	endStr := c.Query("end")
	granularity := c.DefaultQuery("granularity", "5m")

	start, _ := time.Parse(time.RFC3339, startStr)
	end, _ := time.Parse(time.RFC3339, endStr)

	if start.IsZero() {
		end = time.Now()
		start = end.Add(-24 * time.Hour)
	}

	data := h.service.GetResponseTrend(start, end, granularity)
	response.Success(c, data)
}

func (h *MetricsHandler) GetSlowRequestTrend(c *gin.Context) {
	startStr := c.Query("start")
	endStr := c.Query("end")
	granularity := c.DefaultQuery("granularity", "5m")

	start, _ := time.Parse(time.RFC3339, startStr)
	end, _ := time.Parse(time.RFC3339, endStr)

	if start.IsZero() {
		end = time.Now()
		start = end.Add(-24 * time.Hour)
	}

	data := h.service.GetSlowRequestTrend(start, end, granularity)
	response.Success(c, data)
}

func (h *MetricsHandler) GetMethodDistribution(c *gin.Context) {
	startStr := c.Query("start")
	endStr := c.Query("end")

	start, _ := time.Parse(time.RFC3339, startStr)
	end, _ := time.Parse(time.RFC3339, endStr)

	if start.IsZero() {
		end = time.Now()
		start = end.Add(-24 * time.Hour)
	}

	data := h.service.GetMethodDistribution(start, end)
	response.Success(c, data)
}

func (h *MetricsHandler) GetStatusDistribution(c *gin.Context) {
	startStr := c.Query("start")
	endStr := c.Query("end")

	start, _ := time.Parse(time.RFC3339, startStr)
	end, _ := time.Parse(time.RFC3339, endStr)

	if start.IsZero() {
		end = time.Now()
		start = end.Add(-24 * time.Hour)
	}

	data := h.service.GetStatusDistribution(start, end)
	response.Success(c, data)
}

func (h *MetricsHandler) GetErrorRateTrend(c *gin.Context) {
	startStr := c.Query("start")
	endStr := c.Query("end")
	granularity := c.DefaultQuery("granularity", "5m")

	start, _ := time.Parse(time.RFC3339, startStr)
	end, _ := time.Parse(time.RFC3339, endStr)

	if start.IsZero() {
		end = time.Now()
		start = end.Add(-24 * time.Hour)
	}

	data := h.service.GetErrorRateTrend(start, end, granularity)
	response.Success(c, data)
}

func (h *MetricsHandler) GetErrorPaths(c *gin.Context) {
	startStr := c.Query("start")
	endStr := c.Query("end")

	start, _ := time.Parse(time.RFC3339, startStr)
	end, _ := time.Parse(time.RFC3339, endStr)

	if start.IsZero() {
		end = time.Now()
		start = end.Add(-24 * time.Hour)
	}

	data := h.service.GetErrorPaths(start, end, 20)
	response.Success(c, data)
}

func (h *MetricsHandler) GetClientAnalysis(c *gin.Context) {
	startStr := c.Query("start")
	endStr := c.Query("end")

	start, _ := time.Parse(time.RFC3339, startStr)
	end, _ := time.Parse(time.RFC3339, endStr)

	if start.IsZero() {
		end = time.Now()
		start = end.Add(-24 * time.Hour)
	}

	data := h.service.GetClientAnalysis(start, end)
	response.Success(c, data)
}
