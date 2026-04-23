package handler

import (
	"log"
	"nginxops/internal/config"
	"nginxops/internal/database"
	"nginxops/internal/service"
	"nginxops/pkg/response"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type SetupHandler struct {
	setupService *service.SetupService
	authService  *service.AuthService
}

func NewSetupHandler() *SetupHandler {
	return &SetupHandler{
		setupService: service.NewSetupService(),
		authService:  service.NewAuthService(),
	}
}

func (h *SetupHandler) CheckSetupStatus(c *gin.Context) {
	isConfigured := h.setupService.IsConfigured()
	response.Success(c, gin.H{
		"configured": isConfigured,
	})
}

func (h *SetupHandler) InitializeSystem(c *gin.Context) {
	var req service.SetupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if h.setupService.IsConfigured() {
		response.Error(c, 400, "系统已完成初始化")
		return
	}

	if err := h.setupService.InitializeSystem(&req); err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	if err := config.LoadConfig(); err != nil {
		log.Printf("Failed to reload config: %v", err)
		response.Error(c, 500, "配置加载失败: "+err.Error())
		return
	}

	if err := database.InitDB(); err != nil {
		log.Printf("Failed to connect database: %v", err)
		response.Error(c, 500, "数据库连接失败: "+err.Error())
		return
	}

	if err := database.RunMigrations(); err != nil {
		log.Printf("Failed to run migrations: %v", err)
		response.Error(c, 500, "数据库迁移失败: "+err.Error())
		return
	}

	if err := createAdminUser(req.AdminUsername, req.AdminEmail, req.AdminPassword); err != nil {
		log.Printf("Failed to create admin user: %v", err)
		response.Error(c, 500, "创建管理员用户失败: "+err.Error())
		return
	}

	log.Println("System initialization completed successfully!")

	response.SuccessWithMessage(c, "系统初始化成功", gin.H{
		"success": true,
		"message": "系统初始化完成，请使用管理员账号登录",
	})

	go func() {
		time.Sleep(1 * time.Second)
		log.Println("Exiting to restart service in main mode...")
		os.Exit(0)
	}()
}

func createAdminUser(username, email, password string) error {
	var count int64
	database.DB.Model(&struct {
		ID uint `gorm:"primaryKey"`
	}{}).Table("users").Where("username = ?", username).Count(&count)

	if count > 0 {
		log.Printf("Admin user '%s' already exists", username)
		return nil
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	result := database.DB.Exec(`
		INSERT INTO users (username, password, email, role, enabled, created_at, updated_at)
		VALUES (?, ?, ?, 'admin', true, NOW(), NOW())
	`, username, string(hashedPassword), email)

	return result.Error
}
