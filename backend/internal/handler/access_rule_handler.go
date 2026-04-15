package handler

import (
	"nginxops/internal/service"
	"nginxops/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AccessRuleHandler struct {
	svc *service.AccessRuleService
}

func NewAccessRuleHandler() *AccessRuleHandler {
	return &AccessRuleHandler{
		svc: service.NewAccessRuleService(),
	}
}

// ==================== 规则 CRUD ====================

// ListRules 获取访问规则列表
// @Summary 获取访问规则列表
// @Tags 访问规则
// @Produce json
// @Success 200 {object} response.ApiResponse
// @Router /api/access-rules [get]
func (h *AccessRuleHandler) ListRules(c *gin.Context) {
	list, err := h.svc.ListRules()
	if err != nil {
		response.InternalError(c, "获取规则列表失败")
		return
	}
	response.Success(c, list)
}

// GetRule 获取规则详情
// @Summary 获取规则详情
// @Tags 访问规则
// @Param id path int true "规则ID"
// @Produce json
// @Success 200 {object} response.ApiResponse
// @Router /api/access-rules/{id} [get]
func (h *AccessRuleHandler) GetRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的 ID")
		return
	}

	rule, err := h.svc.GetRule(uint(id))
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, rule)
}

// CreateRule 创建访问规则
// @Summary 创建访问规则
// @Tags 访问规则
// @Accept json
// @Produce json
// @Param body body service.AccessRuleDto true "规则内容"
// @Success 200 {object} response.ApiResponse
// @Router /api/access-rules [post]
func (h *AccessRuleHandler) CreateRule(c *gin.Context) {
	var dto service.AccessRuleDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		response.BadRequest(c, "无效的请求数据")
		return
	}

	rule, err := h.svc.CreateRule(&dto)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, rule)
}

// UpdateRule 更新访问规则
// @Summary 更新访问规则
// @Tags 访问规则
// @Accept json
// @Produce json
// @Param id path int true "规则ID"
// @Param body body service.AccessRuleDto true "规则内容"
// @Success 200 {object} response.ApiResponse
// @Router /api/access-rules/{id} [put]
func (h *AccessRuleHandler) UpdateRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的 ID")
		return
	}

	var dto service.AccessRuleDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		response.BadRequest(c, "无效的请求数据")
		return
	}

	rule, err := h.svc.UpdateRule(uint(id), &dto)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, rule)
}

// DeleteRule 删除访问规则
// @Summary 删除访问规则
// @Tags 访问规则
// @Param id path int true "规则ID"
// @Success 200 {object} response.ApiResponse
// @Router /api/access-rules/{id} [delete]
func (h *AccessRuleHandler) DeleteRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的 ID")
		return
	}

	if err := h.svc.DeleteRule(uint(id)); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "已删除"})
}

// ToggleRule 启用/禁用规则
// @Summary 启用/禁用规则
// @Tags 访问规则
// @Param id path int true "规则ID"
// @Param enabled query boolean true "是否启用"
// @Success 200 {object} response.ApiResponse
// @Router /api/access-rules/{id}/toggle [put]
func (h *AccessRuleHandler) ToggleRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的 ID")
		return
	}

	enabled := c.Query("enabled") == "true"
	if err := h.svc.ToggleRule(uint(id), enabled); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "状态已更新"})
}

// ==================== 站点-规则关联 ====================

// GetSiteRules 获取站点关联的规则ID列表
// @Summary 获取站点关联的规则ID列表
// @Tags 访问规则
// @Param siteId path int true "站点ID"
// @Produce json
// @Success 200 {object} response.ApiResponse
// @Router /api/access-rules/sites/{siteId}/rules [get]
func (h *AccessRuleHandler) GetSiteRules(c *gin.Context) {
	siteID, err := strconv.ParseUint(c.Param("siteId"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的站点 ID")
		return
	}

	ruleIDs, err := h.svc.GetSiteRuleIDs(uint(siteID))
	if err != nil {
		response.InternalError(c, "获取站点规则失败")
		return
	}
	response.Success(c, ruleIDs)
}

// SetSiteRules 设置站点关联的规则
// @Summary 设置站点关联的规则
// @Tags 访问规则
// @Accept json
// @Produce json
// @Param siteId path int true "站点ID"
// @Param body body service.SetSiteRulesDto true "规则ID列表"
// @Success 200 {object} response.ApiResponse
// @Router /api/access-rules/sites/{siteId}/rules [put]
func (h *AccessRuleHandler) SetSiteRules(c *gin.Context) {
	siteID, err := strconv.ParseUint(c.Param("siteId"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的站点 ID")
		return
	}

	var dto service.SetSiteRulesDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		response.BadRequest(c, "无效的请求数据")
		return
	}

	if err := h.svc.SetSiteRules(uint(siteID), &dto); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "站点规则已更新"})
}

// ==================== 配置同步 ====================

// SyncConfig 手动同步配置到 Nginx
// @Summary 手动同步配置到 Nginx
// @Tags 访问规则
// @Success 200 {object} response.ApiResponse
// @Router /api/access-rules/sync [post]
func (h *AccessRuleHandler) SyncConfig(c *gin.Context) {
	if err := h.svc.SyncAllConfigs(); err != nil {
		response.InternalError(c, "同步配置失败: "+err.Error())
		return
	}
	response.Success(c, gin.H{"message": "配置已同步到 Nginx"})
}
