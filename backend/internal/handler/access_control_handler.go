package handler

import (
	"net/http"
	"nginxops/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AccessControlHandler struct {
	svc *service.AccessControlService
}

func NewAccessControlHandler() *AccessControlHandler {
	return &AccessControlHandler{
		svc: service.NewAccessControlService(),
	}
}

// ==================== 全局设置 ====================

// GetSettings 获取全局访问控制设置
// @Summary 获取全局访问控制设置
// @Tags 访问控制
// @Produce json
// @Success 200 {object} response.ApiResponse
// @Router /api/access-control/settings [get]
func (h *AccessControlHandler) GetSettings(c *gin.Context) {
	settings, err := h.svc.GetSettings()
	if err != nil {
		InternalServerError(c, "获取设置失败")
		return
	}
	Success(c, settings)
}

// UpdateSettings 更新全局访问控制设置
// @Summary 更新全局访问控制设置
// @Tags 访问控制
// @Accept json
// @Produce json
// @Param body body service.SettingsDto true "设置内容"
// @Success 200 {object} response.ApiResponse
// @Router /api/access-control/settings [put]
func (h *AccessControlHandler) UpdateSettings(c *gin.Context) {
	var dto service.SettingsDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		BadRequest(c, "无效的请求数据")
		return
	}

	if err := h.svc.UpdateSettings(&dto); err != nil {
		InternalServerError(c, "更新设置失败: "+err.Error())
		return
	}

	Success(c, gin.H{"message": "设置已更新"})
}

// ==================== 全局 IP 黑名单 ====================

// ListIPBlacklist 获取 IP 黑名单列表
// @Summary 获取 IP 黑名单列表
// @Tags 访问控制
// @Produce json
// @Success 200 {object} response.ApiResponse
// @Router /api/access-control/ip-blacklist [get]
func (h *AccessControlHandler) ListIPBlacklist(c *gin.Context) {
	list, err := h.svc.ListIPBlacklist()
	if err != nil {
		InternalServerError(c, "获取 IP 黑名单失败")
		return
	}
	Success(c, list)
}

// CreateIPBlacklist 创建 IP 黑名单项
// @Summary 创建 IP 黑名单项
// @Tags 访问控制
// @Accept json
// @Produce json
// @Param body body service.IPBlacklistDto true "黑名单内容"
// @Success 200 {object} response.ApiResponse
// @Router /api/access-control/ip-blacklist [post]
func (h *AccessControlHandler) CreateIPBlacklist(c *gin.Context) {
	var dto service.IPBlacklistDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		BadRequest(c, "无效的请求数据")
		return
	}

	item, err := h.svc.CreateIPBlacklist(&dto)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	Success(c, item)
}

// UpdateIPBlacklist 更新 IP 黑名单项
// @Summary 更新 IP 黑名单项
// @Tags 访问控制
// @Accept json
// @Produce json
// @Param id path int true "黑名单ID"
// @Param body body service.IPBlacklistDto true "黑名单内容"
// @Success 200 {object} response.ApiResponse
// @Router /api/access-control/ip-blacklist/{id} [put]
func (h *AccessControlHandler) UpdateIPBlacklist(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "无效的 ID")
		return
	}

	var dto service.IPBlacklistDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		BadRequest(c, "无效的请求数据")
		return
	}

	item, err := h.svc.UpdateIPBlacklist(uint(id), &dto)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	Success(c, item)
}

// DeleteIPBlacklist 删除 IP 黑名单项
// @Summary 删除 IP 黑名单项
// @Tags 访问控制
// @Param id path int true "黑名单ID"
// @Success 200 {object} response.ApiResponse
// @Router /api/access-control/ip-blacklist/{id} [delete]
func (h *AccessControlHandler) DeleteIPBlacklist(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "无效的 ID")
		return
	}

	if err := h.svc.DeleteIPBlacklist(uint(id)); err != nil {
		InternalServerError(c, "删除失败")
		return
	}

	Success(c, gin.H{"message": "已删除"})
}

// ToggleIPBlacklist 启用/禁用 IP 黑名单项
// @Summary 启用/禁用 IP 黑名单项
// @Tags 访问控制
// @Param id path int true "黑名单ID"
// @Param enabled query boolean true "是否启用"
// @Success 200 {object} response.ApiResponse
// @Router /api/access-control/ip-blacklist/{id}/toggle [put]
func (h *AccessControlHandler) ToggleIPBlacklist(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "无效的 ID")
		return
	}

	enabled := c.Query("enabled") == "true"
	if err := h.svc.ToggleIPBlacklist(uint(id), enabled); err != nil {
		BadRequest(c, err.Error())
		return
	}

	Success(c, gin.H{"message": "状态已更新"})
}

// ==================== 全局 Geo 规则 ====================

// ListGeoRules 获取 Geo 规则列表
// @Summary 获取 Geo 规则列表
// @Tags 访问控制
// @Produce json
// @Success 200 {object} response.ApiResponse
// @Router /api/access-control/geo-rules [get]
func (h *AccessControlHandler) ListGeoRules(c *gin.Context) {
	list, err := h.svc.ListGeoRules()
	if err != nil {
		InternalServerError(c, "获取 Geo 规则失败")
		return
	}
	Success(c, list)
}

// CreateGeoRule 创建 Geo 规则
// @Summary 创建 Geo 规则
// @Tags 访问控制
// @Accept json
// @Produce json
// @Param body body service.GeoRuleDto true "规则内容"
// @Success 200 {object} response.ApiResponse
// @Router /api/access-control/geo-rules [post]
func (h *AccessControlHandler) CreateGeoRule(c *gin.Context) {
	var dto service.GeoRuleDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		BadRequest(c, "无效的请求数据")
		return
	}

	item, err := h.svc.CreateGeoRule(&dto)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	Success(c, item)
}

// UpdateGeoRule 更新 Geo 规则
// @Summary 更新 Geo 规则
// @Tags 访问控制
// @Accept json
// @Produce json
// @Param id path int true "规则ID"
// @Param body body service.GeoRuleDto true "规则内容"
// @Success 200 {object} response.ApiResponse
// @Router /api/access-control/geo-rules/{id} [put]
func (h *AccessControlHandler) UpdateGeoRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "无效的 ID")
		return
	}

	var dto service.GeoRuleDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		BadRequest(c, "无效的请求数据")
		return
	}

	item, err := h.svc.UpdateGeoRule(uint(id), &dto)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	Success(c, item)
}

// DeleteGeoRule 删除 Geo 规则
// @Summary 删除 Geo 规则
// @Tags 访问控制
// @Param id path int true "规则ID"
// @Success 200 {object} response.ApiResponse
// @Router /api/access-control/geo-rules/{id} [delete]
func (h *AccessControlHandler) DeleteGeoRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "无效的 ID")
		return
	}

	if err := h.svc.DeleteGeoRule(uint(id)); err != nil {
		InternalServerError(c, "删除失败")
		return
	}

	Success(c, gin.H{"message": "已删除"})
}

// ToggleGeoRule 启用/禁用 Geo 规则
// @Summary 启用/禁用 Geo 规则
// @Tags 访问控制
// @Param id path int true "规则ID"
// @Param enabled query boolean true "是否启用"
// @Success 200 {object} response.ApiResponse
// @Router /api/access-control/geo-rules/{id}/toggle [put]
func (h *AccessControlHandler) ToggleGeoRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "无效的 ID")
		return
	}

	enabled := c.Query("enabled") == "true"
	if err := h.svc.ToggleGeoRule(uint(id), enabled); err != nil {
		BadRequest(c, err.Error())
		return
	}

	Success(c, gin.H{"message": "状态已更新"})
}

// ==================== 站点专属 IP 黑名单 ====================

// ListSiteIPBlacklist 获取站点 IP 黑名单列表
// @Summary 获取站点 IP 黑名单列表
// @Tags 访问控制
// @Param siteId path int true "站点ID"
// @Produce json
// @Success 200 {object} response.ApiResponse
// @Router /api/access-control/sites/{siteId}/ip-blacklist [get]
func (h *AccessControlHandler) ListSiteIPBlacklist(c *gin.Context) {
	siteID, err := strconv.ParseUint(c.Param("siteId"), 10, 32)
	if err != nil {
		BadRequest(c, "无效的站点 ID")
		return
	}

	list, err := h.svc.ListSiteIPBlacklist(uint(siteID))
	if err != nil {
		InternalServerError(c, "获取站点 IP 黑名单失败")
		return
	}
	Success(c, list)
}

// CreateSiteIPBlacklist 创建站点 IP 黑名单项
// @Summary 创建站点 IP 黑名单项
// @Tags 访问控制
// @Accept json
// @Produce json
// @Param siteId path int true "站点ID"
// @Param body body service.SiteIPBlacklistDto true "黑名单内容"
// @Success 200 {object} response.ApiResponse
// @Router /api/access-control/sites/{siteId}/ip-blacklist [post]
func (h *AccessControlHandler) CreateSiteIPBlacklist(c *gin.Context) {
	siteID, err := strconv.ParseUint(c.Param("siteId"), 10, 32)
	if err != nil {
		BadRequest(c, "无效的站点 ID")
		return
	}

	var dto service.SiteIPBlacklistDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		BadRequest(c, "无效的请求数据")
		return
	}
	dto.SiteID = uint(siteID)

	item, err := h.svc.CreateSiteIPBlacklist(&dto)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	Success(c, item)
}

// DeleteSiteIPBlacklist 删除站点 IP 黑名单项
// @Summary 删除站点 IP 黑名单项
// @Tags 访问控制
// @Param id path int true "黑名单ID"
// @Success 200 {object} response.ApiResponse
// @Router /api/access-control/site-ip-blacklist/{id} [delete]
func (h *AccessControlHandler) DeleteSiteIPBlacklist(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "无效的 ID")
		return
	}

	if err := h.svc.DeleteSiteIPBlacklist(uint(id)); err != nil {
		InternalServerError(c, "删除失败")
		return
	}

	Success(c, gin.H{"message": "已删除"})
}

// ==================== 站点专属 Geo 规则 ====================

// ListSiteGeoRules 获取站点 Geo 规则列表
// @Summary 获取站点 Geo 规则列表
// @Tags 访问控制
// @Param siteId path int true "站点ID"
// @Produce json
// @Success 200 {object} response.ApiResponse
// @Router /api/access-control/sites/{siteId}/geo-rules [get]
func (h *AccessControlHandler) ListSiteGeoRules(c *gin.Context) {
	siteID, err := strconv.ParseUint(c.Param("siteId"), 10, 32)
	if err != nil {
		BadRequest(c, "无效的站点 ID")
		return
	}

	list, err := h.svc.ListSiteGeoRules(uint(siteID))
	if err != nil {
		InternalServerError(c, "获取站点 Geo 规则失败")
		return
	}
	Success(c, list)
}

// CreateSiteGeoRule 创建站点 Geo 规则
// @Summary 创建站点 Geo 规则
// @Tags 访问控制
// @Accept json
// @Produce json
// @Param siteId path int true "站点ID"
// @Param body body service.SiteGeoRuleDto true "规则内容"
// @Success 200 {object} response.ApiResponse
// @Router /api/access-control/sites/{siteId}/geo-rules [post]
func (h *AccessControlHandler) CreateSiteGeoRule(c *gin.Context) {
	siteID, err := strconv.ParseUint(c.Param("siteId"), 10, 32)
	if err != nil {
		BadRequest(c, "无效的站点 ID")
		return
	}

	var dto service.SiteGeoRuleDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		BadRequest(c, "无效的请求数据")
		return
	}
	dto.SiteID = uint(siteID)

	item, err := h.svc.CreateSiteGeoRule(&dto)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	Success(c, item)
}

// DeleteSiteGeoRule 删除站点 Geo 规则
// @Summary 删除站点 Geo 规则
// @Tags 访问控制
// @Param id path int true "规则ID"
// @Success 200 {object} response.ApiResponse
// @Router /api/access-control/site-geo-rules/{id} [delete]
func (h *AccessControlHandler) DeleteSiteGeoRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "无效的 ID")
		return
	}

	if err := h.svc.DeleteSiteGeoRule(uint(id)); err != nil {
		InternalServerError(c, "删除失败")
		return
	}

	Success(c, gin.H{"message": "已删除"})
}

// ==================== 配置同步 ====================

// SyncConfig 手动同步配置到 Nginx
// @Summary 手动同步配置到 Nginx
// @Tags 访问控制
// @Success 200 {object} response.ApiResponse
// @Router /api/access-control/sync [post]
func (h *AccessControlHandler) SyncConfig(c *gin.Context) {
	if err := h.svc.SyncConfig(); err != nil {
		InternalServerError(c, "同步配置失败: "+err.Error())
		return
	}
	Success(c, gin.H{"message": "配置已同步到 Nginx"})
}
