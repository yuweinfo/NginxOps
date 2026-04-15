package handler

import (
	"nginxops/internal/service"

	"github.com/gin-gonic/gin"
)

type AccessCheckHandler struct {
	accessRuleSvc *service.AccessRuleService
}

func NewAccessCheckHandler() *AccessCheckHandler {
	return &AccessCheckHandler{
		accessRuleSvc: service.NewAccessRuleService(),
	}
}

// CheckSiteAccess 检查请求 IP 是否被站点的访问控制规则阻止
// Nginx 通过 auth_request 调用此端点
// 路由: GET /api/access-control/check/site/:siteId
// 返回: 200 = 允许访问, 403 = 拒绝访问
func (h *AccessCheckHandler) CheckSiteAccess(c *gin.Context) {
	siteIDStr := c.Param("siteId")
	if siteIDStr == "" {
		c.Status(200)
		return
	}

	// network_mode: host 下，$remote_addr 即为真实客户端 IP
	// Nginx 通过 X-Real-IP 传递客户端 IP
	clientIP := c.GetHeader("X-Real-IP")
	if clientIP == "" {
		clientIP = c.ClientIP()
	}

	// 查询该站点的合并规则，判断是否阻止
	blocked, reason, err := h.accessRuleSvc.CheckSiteAccessBlocked(clientIP, siteIDStr)
	if err != nil {
		// 查询出错时默认允许（避免误拦截）
		c.Status(200)
		return
	}

	if blocked {
		c.Header("X-Block-Reason", reason)
		c.Status(403)
		return
	}

	c.Status(200)
}
