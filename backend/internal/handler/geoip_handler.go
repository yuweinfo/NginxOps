package handler

import (
	"nginxops/internal/service"
	"nginxops/pkg/response"

	"github.com/gin-gonic/gin"
)

type GeoIpHandler struct {
	service *service.GeoIpService
}

func NewGeoIpHandler() *GeoIpHandler {
	return &GeoIpHandler{
		service: service.NewGeoIpService(),
	}
}

// GetGeo 获取IP地理位置
// GET /api/geo/:ip
func (h *GeoIpHandler) GetGeo(c *gin.Context) {
	ip := c.Param("ip")
	geo := h.service.GetGeo(ip)
	response.Success(c, geo)
}
