package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	JWT      JWTConfig      `yaml:"jwt"`
	Nginx    NginxConfig    `yaml:"nginx"`
	ACME     ACMEConfig     `yaml:"acme"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Name     string `yaml:"name"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	SSLMode  string `yaml:"sslmode"`
}

type JWTConfig struct {
	Secret     string `yaml:"secret"`
	Expiration int64  `yaml:"expiration"`
}

type NginxConfig struct {
	ConfigPath    string `yaml:"config-path"`
	ConfDir       string `yaml:"conf-dir"`
	SSLPath       string `yaml:"ssl-path"`
	AccessLog     string `yaml:"access-log"`
	ReloadCommand string `yaml:"reload-command"`
	TestCommand   string `yaml:"test-command"`
}

type ACMEConfig struct {
	AccountKeyPath string `yaml:"account-key-path"`
}

var AppConfig *Config

var ConfigPath = "/data/config.yml"

func IsConfigured() bool {
	_, err := os.Stat(ConfigPath)
	return err == nil
}

func LoadConfig() error {
	data, err := os.ReadFile(ConfigPath)
	if err != nil {
		return loadFromEnv()
	}

	AppConfig = &Config{}
	if err := yaml.Unmarshal(data, AppConfig); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	setDefaults()

	return nil
}

func loadFromEnv() error {
	AppConfig = &Config{}

	setDefaults()

	if host := os.Getenv("DB_HOST"); host != "" {
		AppConfig.Database.Host = host
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			AppConfig.Database.Port = p
		}
	}
	if name := os.Getenv("DB_NAME"); name != "" {
		AppConfig.Database.Name = name
	}
	if user := os.Getenv("DB_USER"); user != "" {
		AppConfig.Database.User = user
	}
	if password := os.Getenv("DB_PASSWORD"); password != "" {
		AppConfig.Database.Password = password
	}

	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		AppConfig.JWT.Secret = secret
	}

	if dataDir := os.Getenv("DATA_DIR"); dataDir != "" {
		AppConfig.Nginx.ConfDir = fmt.Sprintf("%s/nginx/conf.d", dataDir)
		AppConfig.Nginx.SSLPath = fmt.Sprintf("%s/nginx/ssl", dataDir)
		AppConfig.Nginx.AccessLog = fmt.Sprintf("%s/logs/nginx/access.log", dataDir)
		AppConfig.ACME.AccountKeyPath = fmt.Sprintf("%s/nginx/acme", dataDir)
	}

	return nil
}

func setDefaults() {
	if AppConfig.Server.Port == 0 {
		AppConfig.Server.Port = 8080
	}
	if AppConfig.Database.Host == "" {
		AppConfig.Database.Host = "localhost"
	}
	if AppConfig.Database.Port == 0 {
		AppConfig.Database.Port = 5432
	}
	if AppConfig.Database.Name == "" {
		AppConfig.Database.Name = "nginxops"
	}
	if AppConfig.Database.User == "" {
		AppConfig.Database.User = "nginxops"
	}
	if AppConfig.Database.Password == "" {
		AppConfig.Database.Password = "nginxops123"
	}
	if AppConfig.Database.SSLMode == "" {
		AppConfig.Database.SSLMode = "disable"
	}
	if AppConfig.JWT.Secret == "" {
		AppConfig.JWT.Secret = "your-256-bit-secret-key-here-must-be-at-least-32-characters-long"
	}
	if AppConfig.JWT.Expiration == 0 {
		AppConfig.JWT.Expiration = 86400000
	}
	if AppConfig.Nginx.ConfigPath == "" {
		AppConfig.Nginx.ConfigPath = "/etc/nginx/nginx.conf"
	}
	if AppConfig.Nginx.ConfDir == "" {
		AppConfig.Nginx.ConfDir = "/data/nginx/conf.d"
	}
	if AppConfig.Nginx.SSLPath == "" {
		AppConfig.Nginx.SSLPath = "/data/nginx/ssl"
	}
	if AppConfig.Nginx.AccessLog == "" {
		AppConfig.Nginx.AccessLog = "/data/logs/nginx/access.log"
	}
	if AppConfig.Nginx.ReloadCommand == "" {
		AppConfig.Nginx.ReloadCommand = "nginx -s reload"
	}
	if AppConfig.Nginx.TestCommand == "" {
		AppConfig.Nginx.TestCommand = "nginx -t"
	}
	if AppConfig.ACME.AccountKeyPath == "" {
		AppConfig.ACME.AccountKeyPath = "/data/nginx/acme"
	}
}

func (c *JWTConfig) GetExpirationDuration() time.Duration {
	return time.Duration(c.Expiration) * time.Millisecond
}
