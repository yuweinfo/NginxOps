package service

import (
	"encoding/json"
	"fmt"
	"log"
	"nginxops/internal/config"
	"nginxops/internal/model"
	"nginxops/internal/repository"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type UpstreamService struct {
	repo *repository.UpstreamRepository
}

func NewUpstreamService() *UpstreamService {
	return &UpstreamService{
		repo: repository.NewUpstreamRepository(),
	}
}

type UpstreamDto struct {
	ID            uint              `json:"id"`
	Name          string            `json:"name"`
	LBMode        string            `json:"lbMode"`
	HealthCheck   bool              `json:"healthCheck"`
	CheckInterval int               `json:"checkInterval"`
	CheckPath     string            `json:"checkPath"`
	CheckTimeout  int               `json:"checkTimeout"`
	Servers       []model.UpstreamServer `json:"servers"`
}

// validateUpstream 验证 upstream 配置
func (s *UpstreamService) validateUpstream(dto *UpstreamDto) error {
	// 验证名称格式（只允许字母、数字、下划线）
	if dto.Name == "" {
		return fmt.Errorf("upstream 名称不能为空")
	}
	if matched, _ := regexp.MatchString("^[a-zA-Z][a-zA-Z0-9_]*$", dto.Name); !matched {
		return fmt.Errorf("upstream 名称必须以字母开头，只能包含字母、数字和下划线")
	}

	// 验证服务器配置（如果有服务器的话）
	for i, server := range dto.Servers {
		if server.Host == "" {
			return fmt.Errorf("服务器 %d: 主机地址不能为空", i+1)
		}
		if server.Port <= 0 || server.Port > 65535 {
			return fmt.Errorf("服务器 %d: 端口必须在 1-65535 之间", i+1)
		}
	}

	return nil
}

func (s *UpstreamService) ListAll() ([]UpstreamDto, error) {
	upstreams, err := s.repo.FindAll()
	if err != nil {
		return nil, err
	}

	var dtos []UpstreamDto
	for _, u := range upstreams {
		dtos = append(dtos, s.toDto(&u))
	}
	return dtos, nil
}

func (s *UpstreamService) List(page, size int) ([]model.Upstream, int64, error) {
	return s.repo.FindPage(page, size)
}

func (s *UpstreamService) GetByID(id uint) (*UpstreamDto, error) {
	upstream, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	dto := s.toDto(upstream)
	return &dto, nil
}

func (s *UpstreamService) Create(dto *UpstreamDto) (*UpstreamDto, error) {
	// 验证配置
	if err := s.validateUpstream(dto); err != nil {
		return nil, err
	}

	upstream := &model.Upstream{
		Name:          dto.Name,
		LBMode:        dto.LBMode,
		HealthCheck:   dto.HealthCheck,
		CheckInterval: dto.CheckInterval,
		CheckPath:     dto.CheckPath,
		CheckTimeout:  dto.CheckTimeout,
	}
	upstream.Servers = s.serversToJSON(dto.Servers)

	if err := s.repo.Create(upstream); err != nil {
		return nil, err
	}

	// 写入配置文件
	if err := s.writeConfigFile(upstream); err != nil {
		log.Printf("Warning: 写入 upstream 配置文件失败: %v", err)
	}

	dto = &UpstreamDto{
		ID:            upstream.ID,
		Name:          upstream.Name,
		LBMode:        upstream.LBMode,
		HealthCheck:   upstream.HealthCheck,
		CheckInterval: upstream.CheckInterval,
		CheckPath:     upstream.CheckPath,
		CheckTimeout:  upstream.CheckTimeout,
		Servers:       dto.Servers,
	}
	return dto, nil
}

func (s *UpstreamService) Update(id uint, dto *UpstreamDto) (*UpstreamDto, error) {
	// 验证配置
	if err := s.validateUpstream(dto); err != nil {
		return nil, err
	}

	upstream, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("Upstream 不存在")
	}

	// 保存旧名称，用于删除旧配置文件
	oldName := upstream.Name

	upstream.Name = dto.Name
	upstream.LBMode = dto.LBMode
	upstream.HealthCheck = dto.HealthCheck
	upstream.CheckInterval = dto.CheckInterval
	upstream.CheckPath = dto.CheckPath
	upstream.CheckTimeout = dto.CheckTimeout
	upstream.Servers = s.serversToJSON(dto.Servers)

	if err := s.repo.Update(upstream); err != nil {
		return nil, err
	}

	// 如果名称变更，删除旧配置文件
	if oldName != upstream.Name {
		oldUpstream := &model.Upstream{Name: oldName}
		s.deleteConfigFile(oldUpstream)
	}

	// 写入新配置文件
	if err := s.writeConfigFile(upstream); err != nil {
		log.Printf("Warning: 写入 upstream 配置文件失败: %v", err)
	}

	dto.ID = upstream.ID
	return dto, nil
}

func (s *UpstreamService) Delete(id uint) error {
	upstream, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	// 删除配置文件
	s.deleteConfigFile(upstream)

	return s.repo.Delete(id)
}

func (s *UpstreamService) GenerateConfig(id uint) (string, error) {
	upstream, err := s.repo.FindByID(id)
	if err != nil {
		return "", fmt.Errorf("Upstream 不存在")
	}
	return s.buildUpstreamConfig(upstream), nil
}

func (s *UpstreamService) toDto(entity *model.Upstream) UpstreamDto {
	dto := UpstreamDto{
		ID:            entity.ID,
		Name:          entity.Name,
		LBMode:        entity.LBMode,
		HealthCheck:   entity.HealthCheck,
		CheckInterval: entity.CheckInterval,
		CheckPath:     entity.CheckPath,
		CheckTimeout:  entity.CheckTimeout,
	}

	var servers []model.UpstreamServer
	if err := json.Unmarshal([]byte(entity.Servers), &servers); err == nil {
		dto.Servers = servers
	} else {
		dto.Servers = []model.UpstreamServer{}
	}

	return dto
}

// ToDto 公开方法，供外部调用
func (s *UpstreamService) ToDto(entity *model.Upstream) UpstreamDto {
	return s.toDto(entity)
}

func (s *UpstreamService) serversToJSON(servers []model.UpstreamServer) string {
	if servers == nil {
		return "[]"
	}
	data, err := json.Marshal(servers)
	if err != nil {
		return "[]"
	}
	return string(data)
}

func (s *UpstreamService) buildUpstreamConfig(upstream *model.Upstream) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("upstream %s {\n", upstream.Name))

	// 负载均衡策略
	switch upstream.LBMode {
	case "ip_hash":
		sb.WriteString("    ip_hash;\n")
	case "least_conn":
		sb.WriteString("    least_conn;\n")
	}

	// 健康检查
	if upstream.HealthCheck {
		interval := upstream.CheckInterval
		if interval == 0 {
			interval = 5
		}
		timeout := upstream.CheckTimeout
		if timeout == 0 {
			timeout = 3
		}

		sb.WriteString("    # 健康检查配置（需要 nginx_upstream_check_module）\n")
		sb.WriteString(fmt.Sprintf("    check interval=%dms rise=2 fall=3 timeout=%ds type=http;\n", interval*1000, timeout))

		if upstream.CheckPath != "" {
			sb.WriteString(fmt.Sprintf("    check_http_send \"%s\";\n", upstream.CheckPath))
			sb.WriteString("    check_http_expect_alive http_2xx http_3xx;\n")
		} else {
			sb.WriteString("    check_http_send \"HEAD / HTTP/1.0\\r\\n\\r\\n\";\n")
			sb.WriteString("    check_http_expect_alive http_2xx;\n")
		}
	}

	// 后端服务器
	var servers []model.UpstreamServer
	if err := json.Unmarshal([]byte(upstream.Servers), &servers); err == nil {
		for _, server := range servers {
			sb.WriteString(fmt.Sprintf("    server %s:%d", server.Host, server.Port))

			if server.Weight > 0 {
				sb.WriteString(fmt.Sprintf(" weight=%d", server.Weight))
			}
			if server.MaxFails > 0 {
				sb.WriteString(fmt.Sprintf(" max_fails=%d", server.MaxFails))
			}
			if server.FailTimeout > 0 {
				sb.WriteString(fmt.Sprintf(" fail_timeout=%ds", server.FailTimeout))
			}
			if server.Status == "down" {
				sb.WriteString(" down")
			}
			if server.Backup {
				sb.WriteString(" backup")
			}

			sb.WriteString(";\n")
		}
	}

	sb.WriteString("}\n")
	return sb.String()
}

// writeConfigFile 写入 upstream 配置文件
func (s *UpstreamService) writeConfigFile(upstream *model.Upstream) error {
	confDir := config.AppConfig.Nginx.ConfDir
	if confDir == "" {
		confDir = "/data/nginx/conf.d"
	}

	// 确保目录存在
	if err := os.MkdirAll(confDir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %v", err)
	}

	// 生成配置内容
	configContent := s.buildUpstreamConfig(upstream)

	// 写入文件：upstream_{name}.conf
	fileName := fmt.Sprintf("upstream_%s.conf", upstream.Name)
	filePath := filepath.Join(confDir, fileName)

	if err := os.WriteFile(filePath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	log.Printf("Upstream 配置文件已写入: %s", filePath)
	return nil
}

// deleteConfigFile 删除 upstream 配置文件
func (s *UpstreamService) deleteConfigFile(upstream *model.Upstream) error {
	confDir := config.AppConfig.Nginx.ConfDir
	if confDir == "" {
		confDir = "/data/nginx/conf.d"
	}

	fileName := fmt.Sprintf("upstream_%s.conf", upstream.Name)
	filePath := filepath.Join(confDir, fileName)

	if _, err := os.Stat(filePath); err == nil {
		if err := os.Remove(filePath); err != nil {
			return fmt.Errorf("删除配置文件失败: %v", err)
		}
		log.Printf("Upstream 配置文件已删除: %s", filePath)
	}
	return nil
}
