package handler

import (
	"nginxops/internal/service"
	"nginxops/pkg/response"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type AuditHandler struct {
	service *service.AuditService
}

func NewAuditHandler() *AuditHandler {
	return &AuditHandler{
		service: service.NewAuditService(),
	}
}

// List 查询审计日志
// GET /api/audit
func (h *AuditHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	module := c.Query("module")
	action := c.Query("action")
	status := c.Query("status")
	userID, _ := strconv.ParseUint(c.Query("userId"), 10, 32)

	var startTime, endTime *time.Time
	if startStr := c.Query("startTime"); startStr != "" {
		t, err := time.Parse(time.RFC3339, startStr)
		if err == nil {
			startTime = &t
		}
	}
	if endStr := c.Query("endTime"); endStr != "" {
		t, err := time.Parse(time.RFC3339, endStr)
		if err == nil {
			endTime = &t
		}
	}

	result, err := h.service.Query(page, size, module, action, uint(userID), status, startTime, endTime)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, result)
}

// GetByID 获取审计日志详情
// GET /api/audit/:id
func (h *AuditHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	log, err := h.service.GetByID(uint(id))
	if err != nil || log == nil {
		response.Error(c, 200, "日志不存在")
		return
	}
	response.Success(c, log)
}

// GetModules 获取模块列表
// GET /api/audit/modules
func (h *AuditHandler) GetModules(c *gin.Context) {
	modules := []string{"AUTH", "SITE", "CERTIFICATE", "UPSTREAM", "USER", "CONFIG"}
	response.Success(c, modules)
}

// GetActions 获取操作列表
// GET /api/audit/actions
func (h *AuditHandler) GetActions(c *gin.Context) {
	actions := []string{"LOGIN", "LOGOUT", "PASSWORD_CHANGE", "CREATE", "UPDATE", "DELETE", "ENABLE", "DISABLE", "RELOAD", "TEST", "SYNC", "RENEW", "IMPORT"}
	response.Success(c, actions)
}
