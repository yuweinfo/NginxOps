package handler

import (
	"nginxops/internal/model"
	"nginxops/internal/service"
	"nginxops/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CertificateHandler struct {
	service *service.CertificateService
}

func NewCertificateHandler() *CertificateHandler {
	return &CertificateHandler{
		service: service.NewCertificateService(),
	}
}

// List 获取所有证书
// GET /api/certificates
func (h *CertificateHandler) List(c *gin.Context) {
	certs, err := h.service.ListAll()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, certs)
}

// ListPage 分页查询
// GET /api/certificates/page
func (h *CertificateHandler) ListPage(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	status := c.Query("status")

	certs, total, err := h.service.List(page, size, status)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{
		"records": certs,
		"total":   total,
	})
}

// ListAvailable 获取可用证书
// GET /api/certificates/available
func (h *CertificateHandler) ListAvailable(c *gin.Context) {
	certs, err := h.service.ListAvailable()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, certs)
}

// GetByID 获取证书详情
// GET /api/certificates/:id
func (h *CertificateHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	cert, err := h.service.GetByID(uint(id))
	if err != nil || cert == nil {
		response.Error(c, 200, "证书不存在")
		return
	}
	response.Success(c, cert)
}

// Create 创建证书记录
// POST /api/certificates
func (h *CertificateHandler) Create(c *gin.Context) {
	var req struct {
		Domain string `json:"domain" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	cert := &model.Certificate{Domain: req.Domain}
	result, err := h.service.Create(cert)
	if err != nil {
		response.Error(c, 200, err.Error())
		return
	}
	response.Success(c, result)
}

// Update 更新证书
// PUT /api/certificates/:id
func (h *CertificateHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	var cert model.Certificate
	if err := c.ShouldBindJSON(&cert); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	result, err := h.service.Update(uint(id), &cert)
	if err != nil {
		response.Error(c, 200, err.Error())
		return
	}
	response.Success(c, result)
}

// Delete 删除证书
// DELETE /api/certificates/:id
func (h *CertificateHandler) Delete(c *gin.Context) {
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

// RequestCertificate 申请证书
// POST /api/certificates/:id/request
func (h *CertificateHandler) RequestCertificate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	var req struct {
		Issuer        string `json:"issuer"`
		DNSProviderID uint   `json:"dnsProviderId"`
	}
	c.ShouldBindJSON(&req)

	if req.Issuer == "" {
		req.Issuer = "letsencrypt"
	}

	if err := h.service.RequestCertificate(uint(id), req.Issuer, req.DNSProviderID); err != nil {
		response.Error(c, 200, err.Error())
		return
	}
	response.SuccessWithMessage(c, "证书申请已提交", nil)
}

// Renew 续期证书
// POST /api/certificates/:id/renew
func (h *CertificateHandler) Renew(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	cert, err := h.service.Renew(uint(id))
	if err != nil {
		response.Error(c, 200, err.Error())
		return
	}
	response.Success(c, cert)
}

// ToggleAutoRenew 切换自动续期
// PUT /api/certificates/:id/auto-renew
func (h *CertificateHandler) ToggleAutoRenew(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	autoRenew := c.Query("autoRenew") == "true"
	if err := h.service.ToggleAutoRenew(uint(id), autoRenew); err != nil {
		response.Error(c, 200, err.Error())
		return
	}
	response.SuccessWithMessage(c, "更新成功", nil)
}

// Import 导入证书
// POST /api/certificates/import
func (h *CertificateHandler) Import(c *gin.Context) {
	var dto service.CertificateImportDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	cert, err := h.service.ImportCertificate(&dto)
	if err != nil {
		response.Error(c, 200, err.Error())
		return
	}
	response.Success(c, cert)
}
