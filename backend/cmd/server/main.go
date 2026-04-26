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

func main() {
	if !config.IsConfigured() {
		log.Println("Config file not found, entering setup mode...")
		config.LoadConfig()
		startSetupServer()
		return
	}

	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := database.InitDB(); err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	if err := database.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	authService := service.NewAuthService()
	if err := authService.InitDefaultUser(); err != nil {
		log.Printf("Warning: Failed to init default user: %v", err)
	}

	service.GetLogCollector().Start()

	startMainServer()
}

func startSetupServer() {
	r := gin.Default()
	r.Use(middleware.CORS())
	r.Use(gin.Recovery())

	setupHandler := handler.NewSetupHandler()
	dbTestHandler := handler.NewDBTestHandler()

	api := r.Group("/api")
	{
		api.GET("/health", handler.HealthCheck)

		setup := api.Group("/setup")
		{
			setup.GET("/status", setupHandler.CheckSetupStatus)
			setup.POST("/init", setupHandler.InitializeSystem)
			setup.POST("/test-db", dbTestHandler.TestConnection)
		}
	}

	r.NoRoute(func(c *gin.Context) {
		response.NotFound(c, "Not found")
	})

	log.Println("Setup mode - Server starting on port 8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func startMainServer() {
	r := gin.Default()
	r.Use(middleware.CORS())
	r.Use(gin.Recovery())
	r.Use(middleware.AuditMiddleware())

	authHandler := handler.NewAuthHandler()
	userHandler := handler.NewUserHandler()
	siteHandler := handler.NewSiteHandler()
	upstreamHandler := handler.NewUpstreamHandler()
	certHandler := handler.NewCertificateHandler()
	nginxHandler := handler.NewNginxHandler()
	statsHandler := handler.NewStatsHandler()
	metricsHandler := handler.NewMetricsHandler()
	auditHandler := handler.NewAuditHandler()
	dnsHandler := handler.NewDnsProviderHandler()
	geoHandler := handler.NewGeoIpHandler()
	setupHandler := handler.NewSetupHandler()
	dbTestHandler := handler.NewDBTestHandler()
	networkHandler := handler.NewNetworkHandler()
	healthHandler := handler.NewHealthHandler()
	accessRuleHandler := handler.NewAccessRuleHandler()
	accessCheckHandler := handler.NewAccessCheckHandler()

	logWsHandler := websocket.NewLogWebSocketHandler()

	api := r.Group("/api")
	{
		api.GET("/health", handler.HealthCheck)
		api.GET("/geo/:ip", geoHandler.GetGeo)

		api.GET("/access-control/check/site/:siteId", accessCheckHandler.CheckSiteAccess)

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

	r.GET("/ws/logs", logWsHandler.HandleWebSocket)

	protected := api.Group("")
	protected.Use(middleware.AuthRequired())
	{
		users := protected.Group("/users")
		{
			users.GET("/me", userHandler.GetCurrentUser)
			users.POST("/verify-password", userHandler.VerifyPassword)
			users.PUT("/me", userHandler.UpdateProfile)
		}

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

		nginx := protected.Group("/nginx")
		{
			nginx.GET("/config", nginxHandler.GetConfig)
			nginx.GET("/config/raw", nginxHandler.GetConfigRaw)
			nginx.POST("/config/save", nginxHandler.SaveConfig)
			nginx.GET("/confd", nginxHandler.ListConfFiles)
			nginx.GET("/confd/:fileName", nginxHandler.GetConfFile)
			nginx.POST("/confd/:fileName", nginxHandler.SaveConfFile)
			nginx.GET("/history", nginxHandler.GetHistory)

			nginx.GET("/status", nginxHandler.GetStatus)
			nginx.POST("/start", nginxHandler.Start)
			nginx.POST("/stop", nginxHandler.Stop)
			nginx.POST("/restart", nginxHandler.Restart)
			nginx.POST("/reload", nginxHandler.Reload)
			nginx.POST("/test", nginxHandler.TestConfig)
		}

		stats := protected.Group("/stats")
		{
			stats.GET("/dashboard", statsHandler.GetDashboard)
			stats.GET("/logs", statsHandler.QueryLogs)
		}

		metrics := protected.Group("/metrics")
		{
			metrics.GET("/overview", metricsHandler.GetOverview)
			metrics.GET("/traffic", metricsHandler.GetTrafficTrend)
			metrics.GET("/response", metricsHandler.GetResponseTrend)
			metrics.GET("/slow-requests", metricsHandler.GetSlowRequestTrend)
			metrics.GET("/method-distribution", metricsHandler.GetMethodDistribution)
			metrics.GET("/status-distribution", metricsHandler.GetStatusDistribution)
			metrics.GET("/error-rate", metricsHandler.GetErrorRateTrend)
			metrics.GET("/error-paths", metricsHandler.GetErrorPaths)
			metrics.GET("/client", metricsHandler.GetClientAnalysis)
		}

		audit := protected.Group("/audit")
		{
			audit.GET("", auditHandler.List)
			audit.GET("/:id", auditHandler.GetByID)
			audit.GET("/modules", auditHandler.GetModules)
			audit.GET("/actions", auditHandler.GetActions)
		}

		dns := protected.Group("/dns-providers")
		{
			dns.GET("", dnsHandler.List)
			dns.GET("/:id", dnsHandler.GetByID)
			dns.POST("", dnsHandler.Create)
			dns.PUT("/:id", dnsHandler.Update)
			dns.DELETE("/:id", dnsHandler.Delete)
			dns.PUT("/:id/default", dnsHandler.SetDefault)
		}

		network := protected.Group("/network")
		{
			network.GET("/info", networkHandler.GetNetworkInfo)
			network.POST("/dns-record", networkHandler.CreateDNSRecord)
		}

		accessRules := protected.Group("/access-rules")
		{
			accessRules.GET("", accessRuleHandler.ListRules)
			accessRules.GET("/:id", accessRuleHandler.GetRule)
			accessRules.POST("", accessRuleHandler.CreateRule)
			accessRules.PUT("/:id", accessRuleHandler.UpdateRule)
			accessRules.DELETE("/:id", accessRuleHandler.DeleteRule)
			accessRules.PUT("/:id/toggle", accessRuleHandler.ToggleRule)

			accessRules.GET("/sites/:siteId/rules", accessRuleHandler.GetSiteRules)
			accessRules.PUT("/sites/:siteId/rules", accessRuleHandler.SetSiteRules)

			accessRules.POST("/sync", accessRuleHandler.SyncConfig)
		}
	}

	r.NoRoute(func(c *gin.Context) {
		response.NotFound(c, "Not found")
	})

	port := config.AppConfig.Server.Port
	if port == 0 {
		port = 8080
	}
	log.Printf("Server starting on port %d...", port)
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
