package handler

import (
	"nginxops/internal/service"
	"nginxops/pkg/response"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler() *UserHandler {
	return &UserHandler{
		userService: service.NewUserService(),
	}
}

// GetCurrentUser 获取当前用户详情
// GET /api/users/me
func (h *UserHandler) GetCurrentUser(c *gin.Context) {
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
		"id":        user.ID,
		"username":  user.Username,
		"email":     user.Email,
		"role":      user.Role,
		"createdAt": user.CreatedAt,
	})
}

// VerifyPassword 验证密码
// POST /api/users/verify-password
func (h *UserHandler) VerifyPassword(c *gin.Context) {
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

	var req struct {
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Password == "" {
		response.Error(c, 200, "密码不能为空")
		return
	}

	valid := h.userService.VerifyPassword(user.ID, req.Password)
	if valid {
		response.Success(c, gin.H{"valid": true})
	} else {
		response.Error(c, 200, "密码错误")
	}
}

// UpdateProfile 更新用户资料
// PUT /api/users/me
func (h *UserHandler) UpdateProfile(c *gin.Context) {
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

	var req struct {
		Email       string `json:"email"`
		Password    string `json:"password"`
		OldPassword string `json:"oldPassword"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	// 如果要修改密码，必须验证原密码
	if req.Password != "" {
		if req.OldPassword == "" {
			response.Error(c, 200, "修改密码需要提供原密码")
			return
		}
		if !h.userService.VerifyPassword(user.ID, req.OldPassword) {
			response.Error(c, 200, "原密码错误")
			return
		}
	}

	updated, err := h.userService.UpdateProfile(user.ID, req.Email, req.Password)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"id":       updated.ID,
		"username": updated.Username,
		"email":    updated.Email,
		"role":     updated.Role,
	})
}
