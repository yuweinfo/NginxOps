package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"nginxops/internal/service"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// AuditConfig 审计配置
type AuditConfig struct {
	Action     string
	Module     string
	TargetType string
}

// auditRules 定义需要审计的路由规则
var auditRules = map[string]map[string]AuditConfig{
	"POST": {
		"/api/auth/login":    {Action: "LOGIN", Module: "AUTH", TargetType: "User"},
		"/api/auth/logout":   {Action: "LOGOUT", Module: "AUTH", TargetType: "User"},
		"/api/sites":         {Action: "CREATE", Module: "SITE", TargetType: "Site"},
		"/api/upstreams":     {Action: "CREATE", Module: "UPSTREAM", TargetType: "Upstream"},
		"/api/certificates":  {Action: "CREATE", Module: "CERTIFICATE", TargetType: "Certificate"},
		"/api/certificates/import": {Action: "IMPORT", Module: "CERTIFICATE", TargetType: "Certificate"},
		"/api/sites/sync":    {Action: "SYNC", Module: "SITE", TargetType: "Config"},
		"/api/nginx/reload":  {Action: "RELOAD", Module: "NGINX", TargetType: "Config"},
		"/api/nginx/test":    {Action: "TEST", Module: "NGINX", TargetType: "Config"},
	},
	"PUT": {
		"/api/users/me":      {Action: "UPDATE", Module: "USER", TargetType: "Profile"},
	},
	"DELETE": {
		// 动态路径需要在中间件中处理
	},
}

// AuditMiddleware 审计日志中间件
func AuditMiddleware() gin.HandlerFunc {
	auditService := service.NewAuditService()

	return func(c *gin.Context) {
		// 只记录需要审计的请求
		method := c.Request.Method
		path := c.Request.URL.Path

		// 检查是否需要审计
		config := getAuditConfig(method, path)
		if config == nil {
			c.Next()
			return
		}

		// 记录开始时间
		startTime := time.Now()

		// 读取请求体（用于记录详情）
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 创建响应写入器
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// 执行请求
		c.Next()

		// 记录审计日志
		status := "SUCCESS"
		if c.Writer.Status() >= 400 {
			status = "FAILURE"
		}

		// 获取用户信息
		var userID uint
		var username string
		if uid, exists := c.Get("userId"); exists {
			userID = uid.(uint)
		}
		if uname, exists := c.Get("username"); exists {
			username = uname.(string)
		}

		// 提取目标ID和名称
		targetID := extractTargetID(path)
		targetName := extractTargetName(blw.body.String())

		// 构建详情
		detail := buildDetail(requestBody, path)

		// 获取客户端IP
		ip := c.ClientIP()
		userAgent := c.Request.UserAgent()

		// 异步记录审计日志
		go auditService.Log(
			userID,
			username,
			config.Action,
			config.Module,
			config.TargetType,
			targetID,
			targetName,
			detail,
			ip,
			userAgent,
			status,
			"",
		)

		_ = startTime // 避免未使用警告
	}
}

// getAuditConfig 获取审计配置
func getAuditConfig(method, path string) *AuditConfig {
	// 检查静态规则
	if rules, ok := auditRules[method]; ok {
		if config, ok := rules[path]; ok {
			return &config
		}
	}

	// 检查动态路径
	if method == "PUT" {
		if strings.HasPrefix(path, "/api/sites/") && strings.HasSuffix(path, "/toggle") {
			return &AuditConfig{Action: "UPDATE", Module: "SITE", TargetType: "Site"}
		}
		if strings.HasPrefix(path, "/api/certificates/") && strings.HasSuffix(path, "/auto-renew") {
			return &AuditConfig{Action: "UPDATE", Module: "CERTIFICATE", TargetType: "Certificate"}
		}
		if strings.HasPrefix(path, "/api/certificates/") && strings.HasSuffix(path, "/renew") {
			return &AuditConfig{Action: "RENEW", Module: "CERTIFICATE", TargetType: "Certificate"}
		}
		if strings.HasPrefix(path, "/api/dns-providers/") && strings.HasSuffix(path, "/default") {
			return &AuditConfig{Action: "UPDATE", Module: "DNS", TargetType: "DnsProvider"}
		}
		if strings.HasPrefix(path, "/api/sites/") {
			return &AuditConfig{Action: "UPDATE", Module: "SITE", TargetType: "Site"}
		}
		if strings.HasPrefix(path, "/api/upstreams/") {
			return &AuditConfig{Action: "UPDATE", Module: "UPSTREAM", TargetType: "Upstream"}
		}
		if strings.HasPrefix(path, "/api/certificates/") {
			return &AuditConfig{Action: "UPDATE", Module: "CERTIFICATE", TargetType: "Certificate"}
		}
		if strings.HasPrefix(path, "/api/dns-providers/") {
			return &AuditConfig{Action: "UPDATE", Module: "DNS", TargetType: "DnsProvider"}
		}
	}

	if method == "DELETE" {
		if strings.HasPrefix(path, "/api/sites/") {
			return &AuditConfig{Action: "DELETE", Module: "SITE", TargetType: "Site"}
		}
		if strings.HasPrefix(path, "/api/upstreams/") {
			return &AuditConfig{Action: "DELETE", Module: "UPSTREAM", TargetType: "Upstream"}
		}
		if strings.HasPrefix(path, "/api/certificates/") {
			return &AuditConfig{Action: "DELETE", Module: "CERTIFICATE", TargetType: "Certificate"}
		}
		if strings.HasPrefix(path, "/api/dns-providers/") {
			return &AuditConfig{Action: "DELETE", Module: "DNS", TargetType: "DnsProvider"}
		}
	}

	if method == "POST" {
		if strings.HasPrefix(path, "/api/certificates/") && strings.HasSuffix(path, "/request") {
			return &AuditConfig{Action: "CREATE", Module: "CERTIFICATE", TargetType: "Certificate"}
		}
	}

	return nil
}

// extractTargetID 从路径中提取目标ID
func extractTargetID(path string) uint {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for i := 1; i < len(parts); i++ {
		part := parts[i]
		// 跳过非数字部分
		if len(part) == 0 || part[0] < '0' || part[0] > '9' {
			continue
		}
		// 尝试解析为ID
		var id uint
		if err := json.Unmarshal([]byte(part), &id); err == nil {
			return id
		}
	}
	return 0
}

// extractTargetName 从响应中提取目标名称
func extractTargetName(responseBody string) string {
	var resp struct {
		Data struct {
			Name     string `json:"name"`
			Domain   string `json:"domain"`
			Username string `json:"username"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(responseBody), &resp); err == nil {
		if resp.Data.Name != "" {
			return resp.Data.Name
		}
		if resp.Data.Domain != "" {
			return resp.Data.Domain
		}
		if resp.Data.Username != "" {
			return resp.Data.Username
		}
	}
	return ""
}

// buildDetail 构建审计详情
func buildDetail(requestBody []byte, path string) string {
	detail := make(map[string]interface{})

	// 过滤敏感信息
	if len(requestBody) > 0 {
		var params map[string]interface{}
		if err := json.Unmarshal(requestBody, &params); err == nil {
			filtered := make(map[string]interface{})
			for k, v := range params {
				if !isSensitiveParam(k) {
					filtered[k] = v
				}
			}
			// 即使过滤后为空，也只记录非敏感字段
			if len(filtered) > 0 {
				detail["params"] = filtered
			} else {
				// 如果所有字段都被过滤，至少记录操作已发生（不含敏感信息）
				detail["params"] = map[string]interface{}{"filtered": "仅包含敏感字段"}
			}
		}
	}

	result, _ := json.Marshal(detail)
	return string(result)
}

// isSensitiveParam 检查是否为敏感参数
func isSensitiveParam(paramName string) bool {
	lower := strings.ToLower(paramName)
	return strings.Contains(lower, "password") ||
		strings.Contains(lower, "secret") ||
		strings.Contains(lower, "token") ||
		strings.Contains(lower, "key")
}

// bodyLogWriter 用于捕获响应体
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *bodyLogWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}
