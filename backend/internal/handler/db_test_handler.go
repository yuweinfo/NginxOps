package handler

import (
	"database/sql"
	"fmt"
	"nginxops/pkg/response"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type DBTestHandler struct{}

func NewDBTestHandler() *DBTestHandler {
	return &DBTestHandler{}
}

type DBTestRequest struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Name     string `json:"name"`
	User     string `json:"user"`
	Password string `json:"password"`
}

// TestConnection 测试数据库连接
// POST /api/setup/test-db
func (h *DBTestHandler) TestConnection(c *gin.Context) {
	var req DBTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	// 构建连接字符串
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable connect_timeout=5",
		req.Host, req.Port, req.User, req.Password, req.Name,
	)

	// 尝试连接
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		response.Error(c, 400, fmt.Sprintf("连接失败: %v", err))
		return
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		response.Error(c, 400, fmt.Sprintf("连接失败: %v", err))
		return
	}

	response.Success(c, gin.H{
		"connected": true,
		"message":   "连接成功",
	})
}
