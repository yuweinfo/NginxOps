package service

import (
	"encoding/json"
	"fmt"
	"log"
	"nginxops/internal/config"
	"nginxops/internal/model"
	"nginxops/internal/repository"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type SiteService struct {
	siteRepo     *repository.SiteRepository
	certRepo     *repository.CertificateRepository
	upstreamRepo *repository.UpstreamRepository
}

func NewSiteService() *SiteService {
	return &SiteService{
		siteRepo:     repository.NewSiteRepository(),
		certRepo:     repository.NewCertificateRepository(),
		upstreamRepo: repository.NewUpstreamRepository(),
	}
}

type SiteDto struct {
	ID                uint   `json:"id"`
	FileName          string `json:"fileName"`
	Domain            string `json:"domain"`
	Port              int    `json:"port"`
	SiteType          string `json:"siteType"`
	RootDir           string `json:"rootDir"`
	Locations         string `json:"locations"`
	UpstreamID        *uint  `json:"upstreamId"`
	UpstreamServers   string `json:"upstreamServers"`
	SSLEnabled        bool   `json:"sslEnabled"`
	CertID            *uint  `json:"certId"`
	ForceHttps        bool   `json:"forceHttps"`
	Gzip              bool   `json:"gzip"`
	Cache             bool   `json:"cache"`
	MaxBodySize       int    `json:"maxBodySize"`
	AccessRuleIDs     []uint `json:"accessRuleIds"`
	Enabled           bool   `json:"enabled"`
	Config            string `json:"config"`
}

type PageResult struct {
	Records []model.Site `json:"records"`
	Total   int64        `json:"total"`
}

func (s *SiteService) ListAll() ([]model.Site, error) {
	return s.siteRepo.FindAll()
}

func (s *SiteService) List(page, size int, keyword string) (*PageResult, error) {
	sites, total, err := s.siteRepo.FindPage(page, size, keyword)
	if err != nil {
		return nil, err
	}
	return &PageResult{Records: sites, Total: total}, nil
}

func (s *SiteService) GetByID(id uint) (*model.Site, error) {
	return s.siteRepo.FindByID(id)
}

func (s *SiteService) Create(dto *SiteDto) (*model.Site, error) {
	site := &model.Site{
		FileName:          dto.FileName,
		Domain:            dto.Domain,
		Port:              dto.Port,
		SiteType:          dto.SiteType,
		RootDir:           dto.RootDir,
		Locations:         dto.Locations,
		UpstreamID:        dto.UpstreamID,
		UpstreamServers:   dto.UpstreamServers,
		SSLEnabled:        dto.SSLEnabled,
		CertID:            dto.CertID,
		ForceHttps:        dto.ForceHttps,
		Gzip:              dto.Gzip,
		Cache:             dto.Cache,
		MaxBodySize:       dto.MaxBodySize,
		Enabled:           true,
	}

	if site.FileName == "" {
		site.FileName = strings.Replace(site.Domain, ".", "_", -1) + ".conf"
	}

	// 设置默认值
	if site.MaxBodySize == 0 {
		site.MaxBodySize = 200
	}

	// 先验证配置语法
	nginxConfig := s.buildNginxConfig(site)
	nginxSvc := NewNginxService()
	if valid, errMsg := nginxSvc.ValidateConfig(nginxConfig); !valid {
		return nil, fmt.Errorf("Nginx 配置语法错误: %s", errMsg)
	}

	if err := s.siteRepo.Create(site); err != nil {
		return nil, err
	}

	// 关联访问规则
	if len(dto.AccessRuleIDs) > 0 {
		accessRuleSvc := NewAccessRuleService()
		if err := accessRuleSvc.SetSiteRules(site.ID, &SetSiteRulesDto{RuleIDs: dto.AccessRuleIDs}); err != nil {
			log.Printf("设置站点访问规则失败: %v", err)
		}
	}

	s.writeConfigAndReload(site)
	return site, nil
}

func (s *SiteService) Update(id uint, dto *SiteDto) (*model.Site, error) {
	site, err := s.siteRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("站点不存在")
	}

	// 保存原始值用于回滚
	originalFileName := site.FileName
	originalDomain := site.Domain
	originalPort := site.Port
	originalSiteType := site.SiteType
	originalRootDir := site.RootDir
	originalLocations := site.Locations
	originalUpstreamID := site.UpstreamID
	originalUpstreamServers := site.UpstreamServers
	originalSSLEnabled := site.SSLEnabled
	originalCertID := site.CertID
	originalForceHttps := site.ForceHttps
	originalGzip := site.Gzip
	originalCache := site.Cache
	originalMaxBodySize := site.MaxBodySize

	if dto.FileName != "" {
		site.FileName = dto.FileName
	}
	if dto.Domain != "" {
		site.Domain = dto.Domain
	}
	if dto.Port != 0 {
		site.Port = dto.Port
	}
	if dto.SiteType != "" {
		site.SiteType = dto.SiteType
	}
	if dto.RootDir != "" {
		site.RootDir = dto.RootDir
	}
	if dto.Locations != "" {
		site.Locations = dto.Locations
	}
	site.UpstreamID = dto.UpstreamID
	if dto.UpstreamServers != "" {
		site.UpstreamServers = dto.UpstreamServers
	}
	site.SSLEnabled = dto.SSLEnabled
	site.CertID = dto.CertID
	site.ForceHttps = dto.ForceHttps
	site.Gzip = dto.Gzip
	site.Cache = dto.Cache
	if dto.MaxBodySize > 0 {
		site.MaxBodySize = dto.MaxBodySize
	}
	if dto.Enabled {
		site.Enabled = true
	}

	// 验证新配置语法
	nginxConfig := s.buildNginxConfig(site)
	nginxSvc := NewNginxService()
	if valid, errMsg := nginxSvc.ValidateConfig(nginxConfig); !valid {
		// 恢复原始值
		site.FileName = originalFileName
		site.Domain = originalDomain
		site.Port = originalPort
		site.SiteType = originalSiteType
		site.RootDir = originalRootDir
		site.Locations = originalLocations
		site.UpstreamID = originalUpstreamID
		site.UpstreamServers = originalUpstreamServers
		site.SSLEnabled = originalSSLEnabled
		site.CertID = originalCertID
		site.ForceHttps = originalForceHttps
		site.Gzip = originalGzip
		site.Cache = originalCache
		site.MaxBodySize = originalMaxBodySize
		return nil, fmt.Errorf("Nginx 配置语法错误: %s", errMsg)
	}

	if err := s.siteRepo.Update(site); err != nil {
		return nil, err
	}

	// 更新访问规则关联
	if dto.AccessRuleIDs != nil {
		accessRuleSvc := NewAccessRuleService()
		if err := accessRuleSvc.SetSiteRules(site.ID, &SetSiteRulesDto{RuleIDs: dto.AccessRuleIDs}); err != nil {
			log.Printf("更新站点访问规则失败: %v", err)
		}
	}

	s.writeConfigAndReload(site)
	return site, nil
}

func (s *SiteService) Delete(id uint) error {
	site, err := s.siteRepo.FindByID(id)
	if err != nil {
		return nil
	}

	if site != nil {
		s.deleteConfigFile(site)
		if err := s.siteRepo.Delete(id); err != nil {
			return err
		}
		s.reloadNginx()
	}
	return nil
}

func (s *SiteService) ToggleEnabled(id uint, enabled bool) error {
	site, err := s.siteRepo.FindByID(id)
	if err != nil {
		return fmt.Errorf("站点不存在")
	}

	if enabled {
		// 启用前验证配置语法
		nginxConfig := s.buildNginxConfig(site)
		nginxSvc := NewNginxService()
		if valid, errMsg := nginxSvc.ValidateConfig(nginxConfig); !valid {
			return fmt.Errorf("Nginx 配置语法错误: %s", errMsg)
		}
	}

	site.Enabled = enabled
	if err := s.siteRepo.Update(site); err != nil {
		return err
	}

	if enabled {
		s.writeConfigAndReload(site)
	} else {
		s.deleteConfigFile(site)
		s.reloadNginx()
	}
	return nil
}

func (s *SiteService) GenerateConfig(id uint) (string, error) {
	site, err := s.siteRepo.FindByID(id)
	if err != nil {
		return "", fmt.Errorf("站点不存在")
	}
	return s.buildNginxConfig(site), nil
}

func (s *SiteService) SyncAllConfigs() error {
	sites, err := s.siteRepo.FindByEnabled(true)
	if err != nil {
		return err
	}

	for _, site := range sites {
		if err := s.writeConfigAndReload(&site); err != nil {
			log.Printf("同步站点配置失败: %s, %v", site.Domain, err)
		} else {
			log.Printf("同步站点配置: %s", site.Domain)
		}
	}
	return nil
}

func (s *SiteService) buildNginxConfig(site *model.Site) string {
	var sb strings.Builder
	siteType := site.SiteType
	if siteType == "" {
		siteType = "proxy"
	}

	// 负载均衡：定义 upstream 块
	// 优先使用关联的 upstream，如果没有则使用站点自己定义的 UpstreamServers
	if siteType == "loadbalance" {
		// 检查是否关联了已定义的 upstream
		if site.UpstreamID != nil {
			// 使用已定义的 upstream，不需要在这里生成配置
			// upstream 配置由 upstream_service 单独管理
		} else if site.UpstreamServers != "" {
			// 使用站点自己定义的后端服务器
			upstreamName := strings.Replace(site.Domain, ".", "_", -1) + "_backend"
			sb.WriteString(fmt.Sprintf("upstream %s {\n", upstreamName))

			var servers []map[string]interface{}
			if err := json.Unmarshal([]byte(site.UpstreamServers), &servers); err == nil {
				for _, server := range servers {
					address, _ := server["address"].(string)
					host, _ := server["host"].(string)
					port := 80
					if p, ok := server["port"]; ok {
						switch v := p.(type) {
						case float64:
							port = int(v)
						case int:
							port = v
						}
					}

					serverAddr := address
					if serverAddr == "" && host != "" {
						serverAddr = fmt.Sprintf("%s:%d", host, port)
					}
					if serverAddr == "" {
						continue
					}

					sb.WriteString(fmt.Sprintf("    server %s", serverAddr))
					if weight, ok := server["weight"]; ok {
						w := 1
						switch v := weight.(type) {
						case float64:
							w = int(v)
						case int:
							w = v
						}
						if w != 1 {
							sb.WriteString(fmt.Sprintf(" weight=%d", w))
						}
					}
					if backup, ok := server["backup"].(bool); ok && backup {
						sb.WriteString(" backup")
					}
					sb.WriteString(";\n")
				}
			}
			sb.WriteString("}\n\n")
		}
	}

	// HTTP 重定向到 HTTPS
	if site.SSLEnabled && site.ForceHttps {
		sb.WriteString("server {\n")
		sb.WriteString("    listen 80;\n")
		sb.WriteString(fmt.Sprintf("    server_name %s;\n", site.Domain))
		sb.WriteString("    return 301 https://$server_name$request_uri;\n")
		sb.WriteString("}\n\n")
	}

	// 主 server 块
	sb.WriteString("server {\n")

	// 监听端口
	if site.SSLEnabled {
		sb.WriteString("    listen 443 ssl http2;\n")
	} else {
		port := site.Port
		if port == 0 {
			port = 80
		}
		sb.WriteString(fmt.Sprintf("    listen %d;\n", port))
	}

	sb.WriteString(fmt.Sprintf("    server_name %s;\n\n", site.Domain))

	// 最大上传大小配置
	maxBodySize := site.MaxBodySize
	if maxBodySize == 0 {
		maxBodySize = 200 // 默认200MB
	}
	sb.WriteString(fmt.Sprintf("    client_max_body_size %dm;\n\n", maxBodySize))

	// SSL 配置
	if site.SSLEnabled {
		certPath := ""
		keyPath := ""

		if site.CertID != nil {
			cert, err := s.certRepo.FindByID(*site.CertID)
			if err == nil && cert != nil {
				certPath = cert.CertPath
				keyPath = cert.KeyPath
			}
		}

		if certPath == "" {
			certPath = "/data/nginx/ssl/" + site.Domain + ".crt"
		}
		if keyPath == "" {
			keyPath = "/data/nginx/ssl/" + site.Domain + ".key"
		}

		sb.WriteString(fmt.Sprintf("    ssl_certificate %s;\n", certPath))
		sb.WriteString(fmt.Sprintf("    ssl_certificate_key %s;\n", keyPath))
		sb.WriteString("    ssl_protocols TLSv1.2 TLSv1.3;\n")
		sb.WriteString("    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256;\n")
		sb.WriteString("    ssl_prefer_server_ciphers off;\n")
		sb.WriteString("    ssl_session_cache shared:SSL:10m;\n")
		sb.WriteString("    ssl_session_timeout 1d;\n\n")
	}

	// Gzip 配置
	if site.Gzip {
		sb.WriteString("    gzip on;\n")
		sb.WriteString("    gzip_types text/plain text/css application/json application/javascript text/xml application/xml;\n")
		sb.WriteString("    gzip_min_length 1000;\n\n")
	}

	// 访问控制配置
	accessControlConfig := s.buildAccessControlConfig(site)
	if accessControlConfig != "" {
		sb.WriteString(accessControlConfig)
		sb.WriteString("\n")
	}

	// 根据站点类型生成配置
	switch siteType {
	case "static":
		s.buildStaticConfig(&sb, site)
	case "proxy":
		s.buildProxyConfig(&sb, site)
	case "loadbalance":
		s.buildLoadBalanceConfig(&sb, site)
	}

	sb.WriteString("}\n")
	return sb.String()
}

func (s *SiteService) buildStaticConfig(sb *strings.Builder, site *model.Site) {
	if site.RootDir != "" {
		sb.WriteString(fmt.Sprintf("    root %s;\n", site.RootDir))
		sb.WriteString("    index index.html index.htm;\n")

		if site.Cache {
			sb.WriteString("\n    location ~* \\.(jpg|jpeg|png|gif|ico|css|js|svg|woff|woff2)$ {\n")
			sb.WriteString("        expires 30d;\n")
			sb.WriteString("        add_header Cache-Control \"public, immutable\";\n")
			sb.WriteString("    }\n")
		}
	}

	sb.WriteString("\n    location / {\n")
	sb.WriteString("        try_files $uri $uri/ =404;\n")
	sb.WriteString("    }\n")
}

func (s *SiteService) buildProxyConfig(sb *strings.Builder, site *model.Site) {
	if site.Locations != "" {
		var locations []map[string]interface{}
		if err := json.Unmarshal([]byte(site.Locations), &locations); err == nil {
			for _, loc := range locations {
				path, _ := loc["path"].(string)
				proxyPass, _ := loc["proxyPass"].(string)
				if path == "" || proxyPass == "" {
					continue
				}

				sb.WriteString(fmt.Sprintf("\n    location %s {\n", path))
				sb.WriteString(fmt.Sprintf("        proxy_pass %s;\n", proxyPass))

				if proxyHeaders, ok := loc["proxyHeaders"].(bool); ok && proxyHeaders {
					sb.WriteString("        proxy_set_header Host $host;\n")
					sb.WriteString("        proxy_set_header X-Real-IP $remote_addr;\n")
					sb.WriteString("        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;\n")
					sb.WriteString("        proxy_set_header X-Forwarded-Proto $scheme;\n")
					if site.SSLEnabled {
						sb.WriteString("        proxy_set_header X-Forwarded-Ssl on;\n")
					}
				}

				if websocket, ok := loc["websocket"].(bool); ok && websocket {
					sb.WriteString("        proxy_http_version 1.1;\n")
					sb.WriteString("        proxy_set_header Upgrade $http_upgrade;\n")
					sb.WriteString("        proxy_set_header Connection \"upgrade\";\n")
				}

				sb.WriteString("    }\n")
			}
		}
	}
}

func (s *SiteService) buildLoadBalanceConfig(sb *strings.Builder, site *model.Site) {
	var upstreamName string

	// 如果关联了已定义的 upstream，使用其名称
	if site.UpstreamID != nil {
		upstream, err := s.upstreamRepo.FindByID(*site.UpstreamID)
		if err == nil && upstream != nil {
			upstreamName = upstream.Name
		}
	}

	// 否则使用默认的名称
	if upstreamName == "" {
		upstreamName = strings.Replace(site.Domain, ".", "_", -1) + "_backend"
	}

	sb.WriteString("\n    location / {\n")
	sb.WriteString(fmt.Sprintf("        proxy_pass http://%s;\n", upstreamName))
	sb.WriteString("        proxy_set_header Host $host;\n")
	sb.WriteString("        proxy_set_header X-Real-IP $remote_addr;\n")
	sb.WriteString("        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;\n")
	sb.WriteString("        proxy_set_header X-Forwarded-Proto $scheme;\n")
	sb.WriteString("    }\n")
}

func (s *SiteService) writeConfigAndReload(site *model.Site) error {
	if !site.Enabled {
		return nil
	}

	nginxConfig := s.buildNginxConfig(site)
	confDir := config.AppConfig.Nginx.ConfDir
	fileName := site.FileName

	confPath := filepath.Join(confDir, fileName)
	if err := os.MkdirAll(filepath.Dir(confPath), 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	if err := os.WriteFile(confPath, []byte(nginxConfig), 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	log.Printf("配置文件已写入: %s", confPath)

	// 保存配置到数据库
	site.Config = nginxConfig
	if err := s.siteRepo.Update(site); err != nil {
		return err
	}

	// 重载Nginx
	s.reloadNginx()
	return nil
}

func (s *SiteService) deleteConfigFile(site *model.Site) {
	confDir := config.AppConfig.Nginx.ConfDir
	fileName := site.FileName
	confPath := filepath.Join(confDir, fileName)

	if _, err := os.Stat(confPath); err == nil {
		if err := os.Remove(confPath); err != nil {
			log.Printf("删除配置文件失败: %v", err)
		} else {
			log.Printf("配置文件已删除: %s", confPath)
		}
	}
}

func (s *SiteService) reloadNginx() {
	testCmd := config.AppConfig.Nginx.TestCommand
	reloadCmd := config.AppConfig.Nginx.ReloadCommand

	// 先测试配置
	test := exec.Command("sh", "-c", testCmd)
	if err := test.Run(); err != nil {
		log.Printf("Nginx配置测试失败: %v", err)
		return
	}

	// 重载配置
	reload := exec.Command("sh", "-c", reloadCmd)
	if err := reload.Run(); err != nil {
		log.Printf("Nginx重载失败: %v", err)
	} else {
		log.Println("Nginx配置已重载")
	}
}

// buildAccessControlConfig 生成站点访问控制配置片段
func (s *SiteService) buildAccessControlConfig(site *model.Site) string {
	// 使用新的访问规则服务
	accessRuleSvc := NewAccessRuleService()

	config, err := accessRuleSvc.GetSiteAccessControlConfig(site.ID)
	if err != nil {
		log.Printf("获取访问控制配置失败: %v", err)
		return ""
	}

	return config
}
