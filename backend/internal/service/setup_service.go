package service

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// SetupConfig 初始化配置结构
type SetupConfig struct {
	Database  DatabaseSetupConfig `yaml:"database"`
	JWT       JWTSetupConfig      `yaml:"jwt"`
	SetupDone bool                `yaml:"setupDone"`
}

type DatabaseSetupConfig struct {
	UseExternal bool   `yaml:"useExternal"`
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
	Name        string `yaml:"name"`
	User        string `yaml:"user"`
	Password    string `yaml:"password"`
}

type JWTSetupConfig struct {
	Secret string `yaml:"secret"`
}

type SetupService struct {
	configPath string
}

func NewSetupService() *SetupService {
	return &SetupService{
		configPath: "/data/config.yml",
	}
}

// IsConfigured 检查系统是否已完成初始化配置
func (s *SetupService) IsConfigured() bool {
	// 检查配置文件是否存在
	if _, err := os.Stat(s.configPath); os.IsNotExist(err) {
		return false
	}

	// 读取配置文件检查 setupDone 标志
	config, err := s.LoadConfig()
	if err != nil {
		return false
	}

	return config.SetupDone
}

// LoadConfig 加载配置文件
func (s *SetupService) LoadConfig() (*SetupConfig, error) {
	data, err := ioutil.ReadFile(s.configPath)
	if err != nil {
		return nil, err
	}

	var config SetupConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveConfig 保存配置文件
func (s *SetupService) SaveConfig(config *SetupConfig) error {
	// 确保目录存在
	dir := filepath.Dir(s.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := ioutil.WriteFile(s.configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// InitializeSystem 初始化系统配置
func (s *SetupService) InitializeSystem(req *SetupRequest) error {
	// 验证请求参数
	if err := s.validateSetupRequest(req); err != nil {
		return err
	}

	// 创建配置
	config := &SetupConfig{
		Database: DatabaseSetupConfig{
			UseExternal: req.UseExternalDB,
			Host:        req.DBHost,
			Port:        req.DBPort,
			Name:        req.DBName,
			User:        req.DBUser,
			Password:    req.DBPassword,
		},
		JWT: JWTSetupConfig{
			Secret: req.JWTSecret,
		},
		SetupDone: true,
	}

	// 保存配置
	if err := s.SaveConfig(config); err != nil {
		return err
	}

	return nil
}

// SetupRequest 初始化请求
type SetupRequest struct {
	UseExternalDB bool   `json:"useExternalDB"`
	DBHost        string `json:"dbHost"`
	DBPort        int    `json:"dbPort"`
	DBName        string `json:"dbName"`
	DBUser        string `json:"dbUser"`
	DBPassword    string `json:"dbPassword"`
	JWTSecret     string `json:"jwtSecret"`
	AdminUsername string `json:"adminUsername"`
	AdminEmail    string `json:"adminEmail"`
	AdminPassword string `json:"adminPassword"`
}

func (s *SetupService) validateSetupRequest(req *SetupRequest) error {
	if req.JWTSecret == "" || len(req.JWTSecret) < 32 {
		return errors.New("JWT 密钥必须至少32个字符")
	}

	if req.AdminUsername == "" || len(req.AdminUsername) < 3 {
		return errors.New("管理员用户名必须至少3个字符")
	}

	if req.AdminEmail == "" {
		return errors.New("管理员邮箱不能为空")
	}

	if req.AdminPassword == "" || len(req.AdminPassword) < 6 {
		return errors.New("管理员密码必须至少6个字符")
	}

	if req.UseExternalDB {
		if req.DBHost == "" {
			return errors.New("数据库主机地址不能为空")
		}
		if req.DBPort == 0 {
			return errors.New("数据库端口不能为空")
		}
		if req.DBName == "" {
			return errors.New("数据库名称不能为空")
		}
		if req.DBUser == "" {
			return errors.New("数据库用户名不能为空")
		}
		if req.DBPassword == "" {
			return errors.New("数据库密码不能为空")
		}
	}

	return nil
}

// GetDatabaseConfig 获取数据库配置
func (s *SetupService) GetDatabaseConfig() (*DatabaseSetupConfig, error) {
	config, err := s.LoadConfig()
	if err != nil {
		return nil, err
	}
	return &config.Database, nil
}

// GetJWTConfig 获取 JWT 配置
func (s *SetupService) GetJWTConfig() (*JWTSetupConfig, error) {
	config, err := s.LoadConfig()
	if err != nil {
		return nil, err
	}
	return &config.JWT, nil
}
