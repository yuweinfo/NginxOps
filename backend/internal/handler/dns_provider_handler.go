package handler

import (
	"nginxops/internal/model"
	"nginxops/internal/service"
	"nginxops/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DnsProviderHandler struct {
	service *service.DnsProviderService
}

func NewDnsProviderHandler() *DnsProviderHandler {
	return &DnsProviderHandler{
		service: service.NewDnsProviderService(),
	}
}

// List 获取所有DNS服务商配置
// GET /api/dns-providers
func (h *DnsProviderHandler) List(c *gin.Context) {
	providers, err := h.service.ListAll()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	// 不返回密钥
	for i := range providers {
		providers[i].AccessKeySecret = ""
	}
	response.Success(c, providers)
}

// GetByID 获取DNS服务商配置详情
// GET /api/dns-providers/:id
func (h *DnsProviderHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	provider, err := h.service.GetByID(uint(id))
	if err != nil || provider == nil {
		response.Error(c, 200, "DNS 服务商配置不存在")
		return
	}
	// 不返回密钥
	provider.AccessKeySecret = ""
	response.Success(c, provider)
}

// Create 创建DNS服务商配置
// POST /api/dns-providers
func (h *DnsProviderHandler) Create(c *gin.Context) {
	var provider model.DnsProvider
	if err := c.ShouldBindJSON(&provider); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	result, err := h.service.Create(&provider)
	if err != nil {
		response.Error(c, 200, err.Error())
		return
	}
	result.AccessKeySecret = ""
	response.Success(c, result)
}

// Update 更新DNS服务商配置
// PUT /api/dns-providers/:id
func (h *DnsProviderHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	var provider model.DnsProvider
	if err := c.ShouldBindJSON(&provider); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	result, err := h.service.Update(uint(id), &provider)
	if err != nil {
		response.Error(c, 200, err.Error())
		return
	}
	result.AccessKeySecret = ""
	response.Success(c, result)
}

// Delete 删除DNS服务商配置
// DELETE /api/dns-providers/:id
func (h *DnsProviderHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	if err := h.service.Delete(uint(id)); err != nil {
		response.Error(c, 200, err.Error())
		return
	}
	response.SuccessWithMessage(c, "删除成功", nil)
}

// SetDefault 设置为默认
// PUT /api/dns-providers/:id/default
func (h *DnsProviderHandler) SetDefault(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	if err := h.service.SetDefault(uint(id)); err != nil {
		response.Error(c, 200, err.Error())
		return
	}
	response.SuccessWithMessage(c, "设置成功", nil)
}
