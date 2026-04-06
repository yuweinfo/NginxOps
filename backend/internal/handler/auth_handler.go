package handler

import (
	"nginxops/internal/service"
	"nginxops/pkg/response"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
	userService *service.UserService
}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		authService: service.NewAuthService(),
		userService: service.NewUserService(),
	}
}

// Login 用户登录
// POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	resp, err := h.authService.Login(&req)
	if err != nil {
		response.Error(c, 200, err.Error())
		return
	}

	response.Success(c, resp)
}

// Logout 用户登出
// POST /api/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	response.SuccessWithMessage(c, "登出成功", nil)
}

// GetCurrentUser 获取当前用户信息
// GET /api/auth/me
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	username, exists := c.Get("username")
	if !exists {
		response.Error(c, 200, "未登录")
		return
	}

	user, err := h.userService.GetByUsername(username.(string))
	if err != nil || user == nil {
		response.Error(c, 200, "用户不存在")
		return
	}

	response.Success(c, gin.H{
		"username": user.Username,
		"role":     user.Role,
	})
}
