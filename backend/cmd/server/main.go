package main

import (
	"log"
	"nginxops/internal/config"
	"nginxops/internal/database"
	"nginxops/internal/handler"
	"nginxops/internal/middleware"
	"nginxops/internal/service"
	"nginxops/internal/websocket"
	"nginxops/pkg/response"

	"github.com/gin-gonic/gin"
)

// 初始化模式
var setupMode = false

func main() {
	// 检查配置文件是否存在
	if !config.IsConfigured() {
		log.Println("Config file not found, entering setup mode...")
		setupMode = true
		// 加载默认配置
		config.LoadConfig()
	} else {
		// 加载配置
		if err := config.LoadConfig(); err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}

		// 初始化数据库
		if err := database.InitDB(); err != nil {
			log.Fatalf("Failed to connect database: %v", err)
		}

		// 执行数据库迁移
		if err := database.RunMigrations(); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}

		// 初始化默认用户（使用配置中的管理员账号）
		authService := service.NewAuthService()
		if err := authService.InitDefaultUser(); err != nil {
			log.Printf("Warning: Failed to init default user: %v", err)
		}

		// 启动日志收集服务（单例，内存聚合统计）
		service.GetLogCollector().Start()
	}

	// 创建Gin引擎
	r := gin.Default()

	// 全局中间件
	r.Use(middleware.CORS())
	r.Use(gin.Recovery())
	r.Use(middleware.AuditMiddleware())

	// 初始化Handler
	authHandler := handler.NewAuthHandler()
	userHandler := handler.NewUserHandler()
	siteHandler := handler.NewSiteHandler()
	upstreamHandler := handler.NewUpstreamHandler()
	certHandler := handler.NewCertificateHandler()
	nginxHandler := handler.NewNginxHandler()
	statsHandler := handler.NewStatsHandler()
	auditHandler := handler.NewAuditHandler()
	dnsHandler := handler.NewDnsProviderHandler()
	geoHandler := handler.NewGeoIpHandler()
	setupHandler := handler.NewSetupHandler()
	dbTestHandler := handler.NewDBTestHandler()
	networkHandler := handler.NewNetworkHandler()
	healthHandler := handler.NewHealthHandler()

	// WebSocket Handler
	logWsHandler := websocket.NewLogWebSocketHandler()

	// API路由组
	api := r.Group("/api")
	{
		// 公开接口 - 无需认证
		api.GET("/health", handler.HealthCheck)
		api.GET("/geo/:ip", geoHandler.GetGeo)

		// 初始化接口 - 无需认证
		setup := api.Group("/setup")
		{
			setup.GET("/status", setupHandler.CheckSetupStatus)
			setup.POST("/init", setupHandler.InitializeSystem)
			setup.POST("/test-db", dbTestHandler.TestConnection)
		}

		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", authHandler.Logout)
			auth.GET("/me", middleware.AuthRequired(), authHandler.GetCurrentUser)
		}
	}

	// WebSocket路由
	r.GET("/ws/logs", logWsHandler.HandleWebSocket)

	// 需要认证的路由
	protected := api.Group("")
	protected.Use(middleware.AuthRequired())
	{
		// 用户相关
		users := protected.Group("/users")
		{
			users.GET("/me", userHandler.GetCurrentUser)
			users.POST("/verify-password", userHandler.VerifyPassword)
			users.PUT("/me", userHandler.UpdateProfile)
		}

		// 站点相关
		sites := protected.Group("/sites")
		{
			sites.GET("", siteHandler.List)
			sites.GET("/page", siteHandler.ListPage)
			sites.GET("/:id", siteHandler.GetByID)
			sites.POST("", siteHandler.Create)
			sites.PUT("/:id", siteHandler.Update)
			sites.DELETE("/:id", siteHandler.Delete)
			sites.PUT("/:id/toggle", siteHandler.ToggleEnabled)
			sites.GET("/:id/config", siteHandler.GetConfig)
			sites.POST("/sync", siteHandler.SyncAll)
		}

		// Upstream相关
		upstreams := protected.Group("/upstreams")
		{
			upstreams.GET("", upstreamHandler.List)
			upstreams.GET("/page", upstreamHandler.ListPage)
			upstreams.GET("/:id", upstreamHandler.GetByID)
			upstreams.POST("", upstreamHandler.Create)
			upstreams.PUT("/:id", upstreamHandler.Update)
			upstreams.DELETE("/:id", upstreamHandler.Delete)
			upstreams.GET("/:id/config", upstreamHandler.GetConfig)
			upstreams.POST("/:id/health-check", healthHandler.CheckUpstream)
		}

		// 证书相关
		certs := protected.Group("/certificates")
		{
			certs.GET("", certHandler.List)
			certs.GET("/page", certHandler.ListPage)
			certs.GET("/available", certHandler.ListAvailable)
			certs.GET("/:id", certHandler.GetByID)
			certs.POST("", certHandler.Create)
			certs.POST("/import", certHandler.Import)
			certs.PUT("/:id", certHandler.Update)
			certs.DELETE("/:id", certHandler.Delete)
			certs.POST("/:id/request", certHandler.RequestCertificate)
			certs.POST("/:id/renew", certHandler.Renew)
			certs.PUT("/:id/auto-renew", certHandler.ToggleAutoRenew)
		}

		// Nginx配置和控制
		nginx := protected.Group("/nginx")
		{
			// 配置相关
			nginx.GET("/config", nginxHandler.GetConfig)
			nginx.GET("/config/raw", nginxHandler.GetConfigRaw)
			nginx.POST("/config/save", nginxHandler.SaveConfig)
			nginx.GET("/confd", nginxHandler.ListConfFiles)
			nginx.GET("/confd/:fileName", nginxHandler.GetConfFile)
			nginx.POST("/confd/:fileName", nginxHandler.SaveConfFile)
			nginx.GET("/history", nginxHandler.GetHistory)

			// 控制相关
			nginx.GET("/status", nginxHandler.GetStatus)
			nginx.POST("/start", nginxHandler.Start)
			nginx.POST("/stop", nginxHandler.Stop)
			nginx.POST("/restart", nginxHandler.Restart)
			nginx.POST("/reload", nginxHandler.Reload)
			nginx.POST("/test", nginxHandler.TestConfig)
		}

		// 统计相关
		stats := protected.Group("/stats")
		{
			stats.GET("/dashboard", statsHandler.GetDashboard)
			stats.GET("/logs", statsHandler.QueryLogs)
		}

		// 审计日志
		audit := protected.Group("/audit")
		{
			audit.GET("", auditHandler.List)
			audit.GET("/:id", auditHandler.GetByID)
			audit.GET("/modules", auditHandler.GetModules)
			audit.GET("/actions", auditHandler.GetActions)
		}

		// DNS服务商
		dns := protected.Group("/dns-providers")
		{
			dns.GET("", dnsHandler.List)
			dns.GET("/:id", dnsHandler.GetByID)
			dns.POST("", dnsHandler.Create)
			dns.PUT("/:id", dnsHandler.Update)
			dns.DELETE("/:id", dnsHandler.Delete)
			dns.PUT("/:id/default", dnsHandler.SetDefault)
		}

		// 网络信息
		network := protected.Group("/network")
		{
			network.GET("/info", networkHandler.GetNetworkInfo)
			network.POST("/dns-record", networkHandler.CreateDNSRecord)
		}
	}

	// 404处理
	r.NoRoute(func(c *gin.Context) {
		response.NotFound(c, "Not found")
	})

	// 启动服务
	port := config.AppConfig.Server.Port
	if port == 0 {
		port = 8080
	}
	log.Printf("Server starting on port %d...", port)
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
