package handler

import (
	"nginxops/internal/service"
	"nginxops/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SiteHandler struct {
	siteService *service.SiteService
}

func NewSiteHandler() *SiteHandler {
	return &SiteHandler{
		siteService: service.NewSiteService(),
	}
}

// List 获取所有站点
// GET /api/sites
func (h *SiteHandler) List(c *gin.Context) {
	sites, err := h.siteService.ListAll()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, sites)
}

// ListPage 分页查询站点
// GET /api/sites/page
func (h *SiteHandler) ListPage(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	keyword := c.Query("keyword")

	result, err := h.siteService.List(page, size, keyword)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, result)
}

// GetByID 获取站点详情
// GET /api/sites/:id
func (h *SiteHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	site, err := h.siteService.GetByID(uint(id))
	if err != nil || site == nil {
		response.Error(c, 200, "站点不存在")
		return
	}
	response.Success(c, site)
}

// Create 创建站点
// POST /api/sites
func (h *SiteHandler) Create(c *gin.Context) {
	var dto service.SiteDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	site, err := h.siteService.Create(&dto)
	if err != nil {
		response.Error(c, 200, err.Error())
		return
	}
	response.Success(c, site)
}

// Update 更新站点
// PUT /api/sites/:id
func (h *SiteHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	var dto service.SiteDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	site, err := h.siteService.Update(uint(id), &dto)
	if err != nil {
		response.Error(c, 200, err.Error())
		return
	}
	response.Success(c, site)
}

// Delete 删除站点
// DELETE /api/sites/:id
func (h *SiteHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	if err := h.siteService.Delete(uint(id)); err != nil {
		response.Error(c, 200, err.Error())
		return
	}
	response.SuccessWithMessage(c, "删除成功", nil)
}

// ToggleEnabled 启用/禁用站点
// PUT /api/sites/:id/toggle
func (h *SiteHandler) ToggleEnabled(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	enabled := c.Query("enabled") == "true"
	if err := h.siteService.ToggleEnabled(uint(id), enabled); err != nil {
		response.Error(c, 200, err.Error())
		return
	}

	site, _ := h.siteService.GetByID(uint(id))
	response.Success(c, site)
}

// GetConfig 获取站点Nginx配置
// GET /api/sites/:id/config
func (h *SiteHandler) GetConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	config, err := h.siteService.GenerateConfig(uint(id))
	if err != nil {
		response.Error(c, 200, err.Error())
		return
	}
	response.Success(c, config)
}

// SyncAll 同步所有站点配置
// POST /api/sites/sync
func (h *SiteHandler) SyncAll(c *gin.Context) {
	if err := h.siteService.SyncAllConfigs(); err != nil {
		response.Error(c, 200, err.Error())
		return
	}
	response.SuccessWithMessage(c, "配置同步成功", nil)
}
