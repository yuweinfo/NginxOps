package handler

import (
	"log"
	"nginxops/internal/config"
	"nginxops/internal/database"
	"nginxops/internal/service"
	"nginxops/pkg/response"
	"os"
	"os/exec"
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

// CheckSetupStatus 检查系统初始化状态
// GET /api/setup/status
func (h *SetupHandler) CheckSetupStatus(c *gin.Context) {
	isConfigured := h.setupService.IsConfigured()
	response.Success(c, gin.H{
		"configured": isConfigured,
	})
}

// InitializeSystem 初始化系统
// POST /api/setup/init
func (h *SetupHandler) InitializeSystem(c *gin.Context) {
	var req service.SetupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	// 检查是否已经初始化
	if h.setupService.IsConfigured() {
		response.Error(c, 400, "系统已完成初始化")
		return
	}

	// 初始化系统配置
	if err := h.setupService.InitializeSystem(&req); err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	// 同步执行数据库初始化（确保在返回前完成）
	// 如果使用内部数据库，启动 PostgreSQL
	if !req.UseExternalDB {
		log.Println("Starting internal PostgreSQL...")
		
		// 预先设置目录权限（确保 postgres 用户有权限）
		exec.Command("chown", "-R", "postgres:postgres", "/data/postgresql").Run()
		exec.Command("chmod", "700", "/data/postgresql").Run()
		
		// 检查数据库集群是否已初始化
		initFlag := "/data/postgresql/PG_VERSION"
		if _, err := exec.Command("test", "-f", initFlag).CombinedOutput(); err != nil {
			log.Println("Initializing PostgreSQL database cluster...")
			cmd := exec.Command("su", "-", "postgres", "-c", 
				"initdb -D /data/postgresql --encoding=UTF8 --locale=en_US.UTF-8")
			if output, err := cmd.CombinedOutput(); err != nil {
				log.Printf("initdb output: %s", string(output))
				// 回滚配置文件
				os.Remove("/data/config.yml")
				response.Error(c, 500, "数据库初始化失败: "+string(output))
				return
			}
			
			// 配置 PostgreSQL
			exec.Command("su", "-", "postgres", "-c",
				"echo \"listen_addresses = '127.0.0.1'\" >> /data/postgresql/postgresql.conf").Run()
			exec.Command("su", "-", "postgres", "-c",
				"echo \"port = 5432\" >> /data/postgresql/postgresql.conf").Run()
			exec.Command("su", "-", "postgres", "-c",
				"echo \"host all all 127.0.0.1/32 trust\" >> /data/postgresql/pg_hba.conf").Run()
			log.Println("Database cluster initialized!")
		}

		// 启动 PostgreSQL
		cmd := exec.Command("su", "-", "postgres", "-c", 
			"pg_ctl -D /data/postgresql -l /data/logs/postgresql.log -w start")
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("pg_ctl output: %s", string(output))
			// 回滚配置文件
			os.Remove("/data/config.yml")
			response.Error(c, 500, "PostgreSQL 启动失败: "+string(output))
			return
		}
		log.Println("PostgreSQL started successfully!")

		// 等待 PostgreSQL 完全就绪
		log.Println("Waiting for PostgreSQL to be ready...")
		for i := 0; i < 30; i++ {
			time.Sleep(1 * time.Second)
			cmd := exec.Command("pg_isready", "-h", "127.0.0.1", "-p", "5432")
			if cmd.Run() == nil {
				log.Println("PostgreSQL is ready!")
				break
			}
			if i == 29 {
				// 回滚配置文件
				os.Remove("/data/config.yml")
				response.Error(c, 500, "PostgreSQL 启动超时")
				return
			}
		}

		// 检查数据库是否存在
		checkDB := exec.Command("su", "-", "postgres", "-c",
			"psql -h 127.0.0.1 -d postgres -tAc \"SELECT 1 FROM pg_database WHERE datname = '"+req.DBName+"'\"")
		dbOutput, _ := checkDB.CombinedOutput()
		if string(dbOutput) != "1\n" {
			// 创建数据库和用户
			log.Println("Creating database and user...")
			exec.Command("su", "-", "postgres", "-c",
				"psql -h 127.0.0.1 -d postgres -c \"CREATE USER \\\""+req.DBUser+"\\\" WITH PASSWORD '"+req.DBPassword+"';\"").Run()
			exec.Command("su", "-", "postgres", "-c",
				"psql -h 127.0.0.1 -d postgres -c \"CREATE DATABASE \\\""+req.DBName+"\\\" OWNER \\\""+req.DBUser+"\\\";\"").Run()
			exec.Command("su", "-", "postgres", "-c",
				"psql -h 127.0.0.1 -d postgres -c \"GRANT ALL PRIVILEGES ON DATABASE \\\""+req.DBName+"\\\" TO \\\""+req.DBUser+"\\\";\"").Run()
			log.Println("Database and user created!")
		}
	}

	// 重新加载配置
	if err := config.LoadConfig(); err != nil {
		log.Printf("Failed to reload config: %v", err)
		os.Remove("/data/config.yml")
		response.Error(c, 500, "配置加载失败: "+err.Error())
		return
	}

	// 连接数据库
	if err := database.InitDB(); err != nil {
		log.Printf("Failed to connect database: %v", err)
		os.Remove("/data/config.yml")
		response.Error(c, 500, "数据库连接失败: "+err.Error())
		return
	}

	// 执行数据库迁移
	if err := database.RunMigrations(); err != nil {
		log.Printf("Failed to run migrations: %v", err)
		os.Remove("/data/config.yml")
		response.Error(c, 500, "数据库迁移失败: "+err.Error())
		return
	}

	// 创建管理员用户
	if err := createAdminUser(req.AdminUsername, req.AdminEmail, req.AdminPassword); err != nil {
		log.Printf("Failed to create admin user: %v", err)
		os.Remove("/data/config.yml")
		response.Error(c, 500, "创建管理员用户失败: "+err.Error())
		return
	}

	log.Println("System initialization completed successfully!")

	// 如果使用内置数据库，启动 supervisor 管理的 PostgreSQL
	if !req.UseExternalDB {
		log.Println("Starting PostgreSQL under supervisor control...")
		cmd := exec.Command("supervisorctl", "-c", "/tmp/supervisord.conf", "start", "postgresql")
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("supervisorctl start postgresql warning: %s (output: %s)", err, string(output))
			// 不影响主流程，只是无法由 supervisor 管理
		} else {
			log.Println("PostgreSQL started under supervisor control!")
		}
	}

	// 返回成功响应
	response.SuccessWithMessage(c, "系统初始化成功", gin.H{
		"success": true,
		"message": "系统初始化完成，请使用管理员账号登录",
	})
}

// createAdminUser 创建管理员用户
func createAdminUser(username, email, password string) error {
	// 检查用户是否已存在
	var count int64
	database.DB.Model(&struct {
		ID uint `gorm:"primaryKey"`
	}{}).Table("users").Where("username = ?", username).Count(&count)
	
	if count > 0 {
		log.Printf("Admin user '%s' already exists", username)
		return nil
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 创建用户
	result := database.DB.Exec(`
		INSERT INTO users (username, password, email, role, enabled, created_at, updated_at)
		VALUES (?, ?, ?, 'admin', true, NOW(), NOW())
	`, username, string(hashedPassword), email)

	return result.Error
}
