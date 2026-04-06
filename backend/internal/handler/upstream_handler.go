package handler

import (
	"nginxops/internal/model"
	"nginxops/internal/service"
	"nginxops/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UpstreamHandler struct {
	service *service.UpstreamService
}

func NewUpstreamHandler() *UpstreamHandler {
	return &UpstreamHandler{
		service: service.NewUpstreamService(),
	}
}

// List 获取所有Upstream
// GET /api/upstreams
func (h *UpstreamHandler) List(c *gin.Context) {
	dtos, err := h.service.ListAll()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, dtos)
}

// ListPage 分页查询
// GET /api/upstreams/page
func (h *UpstreamHandler) ListPage(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	if page < 1 {
		response.Error(c, 200, "页码必须大于0")
		return
	}
	if size < 1 || size > 100 {
		response.Error(c, 200, "页大小必须在1-100之间")
		return
	}

	upstreams, total, err := h.service.List(page, size)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{
		"records": upstreams,
		"total":   total,
	})
}

// GetByID 获取详情
// GET /api/upstreams/:id
func (h *UpstreamHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, 200, "无效的ID")
		return
	}

	dto, err := h.service.GetByID(uint(id))
	if err != nil || dto == nil {
		response.Error(c, 200, "Upstream 不存在")
		return
	}
	response.Success(c, dto)
}

// Create 创建Upstream
// POST /api/upstreams
func (h *UpstreamHandler) Create(c *gin.Context) {
	var dto service.UpstreamDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	result, err := h.service.Create(&dto)
	if err != nil {
		response.Error(c, 200, "创建失败: "+err.Error())
		return
	}
	response.Success(c, result)
}

// Update 更新Upstream
// PUT /api/upstreams/:id
func (h *UpstreamHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		response.Error(c, 200, "无效的ID")
		return
	}

	var dto service.UpstreamDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	result, err := h.service.Update(uint(id), &dto)
	if err != nil {
		response.Error(c, 200, err.Error())
		return
	}
	response.Success(c, result)
}

// Delete 删除Upstream
// DELETE /api/upstreams/:id
func (h *UpstreamHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		response.Error(c, 200, "无效的ID")
		return
	}

	// 检查是否存在
	dto, _ := h.service.GetByID(uint(id))
	if dto == nil {
		response.Error(c, 200, "Upstream 不存在")
		return
	}

	if err := h.service.Delete(uint(id)); err != nil {
		response.Error(c, 200, "删除失败: "+err.Error())
		return
	}
	response.SuccessWithMessage(c, "删除成功", nil)
}

// GetConfig 获取Nginx配置
// GET /api/upstreams/:id/config
func (h *UpstreamHandler) GetConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	config, err := h.service.GenerateConfig(uint(id))
	if err != nil {
		response.Error(c, 200, err.Error())
		return
	}
	response.Success(c, config)
}

// convertUpstreamList 转换Upstream列表
func convertUpstreamList(upstreams []model.Upstream) []service.UpstreamDto {
	var result []service.UpstreamDto
	svc := service.NewUpstreamService()
	for _, u := range upstreams {
		dto := svc.ToDto(&u)
		result = append(result, dto)
	}
	return result
}
