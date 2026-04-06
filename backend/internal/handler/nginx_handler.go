package handler

import (
	"nginxops/internal/service"
	"nginxops/pkg/response"
	"strings"

	"github.com/gin-gonic/gin"
)

type NginxHandler struct {
	service *service.NginxService
}

func NewNginxHandler() *NginxHandler {
	return &NginxHandler{
		service: service.NewNginxService(),
	}
}

// GetConfig 获取nginx配置（结构化）
// GET /api/nginx/config
func (h *NginxHandler) GetConfig(c *gin.Context) {
	// 简化版本：返回原始内容
	// TODO: 实现结构化解析
	content, err := h.service.GetMainConfigRaw()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, content)
}

// GetConfigRaw 获取原始nginx配置
// GET /api/nginx/config/raw
func (h *NginxHandler) GetConfigRaw(c *gin.Context) {
	content, err := h.service.GetMainConfigRaw()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	c.String(200, content)
}

// SaveConfig 保存nginx配置
// POST /api/nginx/config/save
func (h *NginxHandler) SaveConfig(c *gin.Context) {
	var req struct {
		Content string `json:"content"`
		Remark  string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if req.Content == "" {
		response.BadRequest(c, "配置内容不能为空")
		return
	}

	result, err := h.service.SaveAndReload("main", req.Content, req.Remark)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, result)
}

// ListConfFiles 列出conf.d文件
// GET /api/nginx/confd
func (h *NginxHandler) ListConfFiles(c *gin.Context) {
	files, err := h.service.ListConfFiles()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, files)
}

// GetConfFile 获取conf.d文件内容
// GET /api/nginx/confd/:fileName
func (h *NginxHandler) GetConfFile(c *gin.Context) {
	fileName := c.Param("fileName")
	if !isValidFileName(fileName) {
		response.BadRequest(c, "无效的文件名")
		return
	}

	content, err := h.service.GetConfFile(fileName)
	if err != nil {
		response.NotFound(c, "文件不存在")
		return
	}
	c.String(200, content)
}

// SaveConfFile 保存conf.d文件
// POST /api/nginx/confd/:fileName
func (h *NginxHandler) SaveConfFile(c *gin.Context) {
	fileName := c.Param("fileName")
	if !isValidFileName(fileName) {
		response.BadRequest(c, "无效的文件名")
		return
	}

	var req struct {
		Content string `json:"content"`
		Remark  string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if req.Content == "" {
		response.BadRequest(c, "配置内容不能为空")
		return
	}

	result, err := h.service.SaveAndReload(fileName, req.Content, req.Remark)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, result)
}

// GetHistory 获取配置历史
// GET /api/nginx/history
func (h *NginxHandler) GetHistory(c *gin.Context) {
	configName := c.Query("configName")
	limit := 20
	histories, err := h.service.GetHistory(configName, limit)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, histories)
}

// GetStatus 获取Nginx状态
// GET /api/nginx/status
func (h *NginxHandler) GetStatus(c *gin.Context) {
	status := h.service.GetStatus()
	response.Success(c, status)
}

// Start 启动Nginx
// POST /api/nginx/start
func (h *NginxHandler) Start(c *gin.Context) {
	msg, err := h.service.Start()
	if err != nil {
		response.Error(c, 200, err.Error())
		return
	}
	response.SuccessWithMessage(c, msg, nil)
}

// Stop 停止Nginx
// POST /api/nginx/stop
func (h *NginxHandler) Stop(c *gin.Context) {
	msg, err := h.service.Stop()
	if err != nil {
		response.Error(c, 200, err.Error())
		return
	}
	response.SuccessWithMessage(c, msg, nil)
}

// Restart 重启Nginx
// POST /api/nginx/restart
func (h *NginxHandler) Restart(c *gin.Context) {
	msg, err := h.service.Restart()
	if err != nil {
		response.Error(c, 200, err.Error())
		return
	}
	response.SuccessWithMessage(c, msg, nil)
}

// Reload 重载配置
// POST /api/nginx/reload
func (h *NginxHandler) Reload(c *gin.Context) {
	msg, err := h.service.Reload()
	if err != nil {
		response.Error(c, 200, err.Error())
		return
	}
	response.SuccessWithMessage(c, msg, nil)
}

// TestConfig 测试配置
// POST /api/nginx/test
func (h *NginxHandler) TestConfig(c *gin.Context) {
	result := h.service.TestConfig()
	response.Success(c, result)
}

func isValidFileName(fileName string) bool {
	if fileName == "" || strings.Contains(fileName, "..") {
		return false
	}
	return strings.HasSuffix(fileName, ".conf")
}
