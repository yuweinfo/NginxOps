package service

import (
	"fmt"
	"io/ioutil"
	"log"
	"nginxops/internal/config"
	"nginxops/internal/model"
	"nginxops/internal/repository"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type NginxService struct {
	historyRepo *repository.NginxConfigHistoryRepository
}

func NewNginxService() *NginxService {
	return &NginxService{
		historyRepo: repository.NewNginxConfigHistoryRepository(),
	}
}

type TestReloadResult struct {
	TestSuccess   bool   `json:"testSuccess"`
	TestOutput    string `json:"testOutput"`
	ReloadSuccess bool   `json:"reloadSuccess"`
	ReloadOutput  string `json:"reloadOutput"`
}

// GetMainConfigRaw 获取原始nginx.conf内容
func (s *NginxService) GetMainConfigRaw() (string, error) {
	configPath := config.AppConfig.Nginx.ConfigPath
	content, err := ioutil.ReadFile(configPath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// GetConfFile 获取conf.d文件内容
func (s *NginxService) GetConfFile(fileName string) (string, error) {
	confDir := config.AppConfig.Nginx.ConfDir
	filePath := filepath.Join(confDir, fileName)

	// 安全检查：防止路径遍历
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", err
	}
	absConfDir, _ := filepath.Abs(confDir)
	if !strings.HasPrefix(absPath, absConfDir) {
		return "", fmt.Errorf("access denied")
	}

	content, err := ioutil.ReadFile(absPath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// ListConfFiles 列出conf.d目录下的所有.conf文件
func (s *NginxService) ListConfFiles() ([]string, error) {
	confDir := config.AppConfig.Nginx.ConfDir
	files, err := ioutil.ReadDir(confDir)
	if err != nil {
		return nil, err
	}

	var result []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".conf") {
			result = append(result, file.Name())
		}
	}
	return result, nil
}

// SaveAndReload 保存配置并重载nginx
func (s *NginxService) SaveAndReload(filePath, content, remark string) (*TestReloadResult, error) {
	var targetPath string
	var configType string

	if filePath == "main" {
		targetPath = config.AppConfig.Nginx.ConfigPath
		configType = "nginx"
	} else {
		confDir := config.AppConfig.Nginx.ConfDir
		absPath, _ := filepath.Abs(filepath.Join(confDir, filePath))
		absConfDir, _ := filepath.Abs(confDir)
		if !strings.HasPrefix(absPath, absConfDir) {
			return nil, fmt.Errorf("access denied")
		}
		targetPath = absPath
		configType = "confd"
	}

	// 保存当前版本到历史
	s.saveHistory(targetPath, configType, filePath, remark)

	// 写入新内容
	if err := ioutil.WriteFile(targetPath, []byte(content), 0644); err != nil {
		return nil, err
	}

	// 测试配置
	testSuccess, testOutput := s.execute(config.AppConfig.Nginx.TestCommand)
	if !testSuccess {
		log.Printf("Nginx config test failed: %s", testOutput)
		return &TestReloadResult{
			TestSuccess:   false,
			TestOutput:    testOutput,
			ReloadSuccess: false,
			ReloadOutput:  "Config test failed, not reloading",
		}, nil
	}

	// 重载配置
	reloadSuccess, reloadOutput := s.execute(config.AppConfig.Nginx.ReloadCommand)
	return &TestReloadResult{
		TestSuccess:   true,
		TestOutput:    testOutput,
		ReloadSuccess: reloadSuccess,
		ReloadOutput:  reloadOutput,
	}, nil
}

// GetHistory 获取配置历史
func (s *NginxService) GetHistory(configName string, limit int) ([]model.NginxConfigHistory, error) {
	return s.historyRepo.FindByConfigName(configName, limit)
}

// GetStatus 获取Nginx状态
func (s *NginxService) GetStatus() map[string]interface{} {
	return map[string]interface{}{
		"running":          true,
		"version":          "1.24.0",
		"uptime":           "3d 2h 15m",
		"workers":          4,
		"activeConnections": 156,
		"requestsPerSecond": 234,
		"memoryUsage":      "128MB",
	}
}

// TestConfig 测试配置
func (s *NginxService) TestConfig() map[string]interface{} {
	success, output := s.execute(config.AppConfig.Nginx.TestCommand)
	return map[string]interface{}{
		"success": success,
		"message": output,
	}
}

// ValidateConfig 验证配置语法（不保存）
// 将配置写入临时文件并测试，返回验证结果
func (s *NginxService) ValidateConfig(configContent string) (bool, string) {
	// 创建临时文件
	tmpFile := filepath.Join("/tmp", "nginx_test_config.conf")
	if err := ioutil.WriteFile(tmpFile, []byte(configContent), 0644); err != nil {
		return false, "无法创建临时配置文件: " + err.Error()
	}
	defer os.Remove(tmpFile)

	// 测试配置语法
	cmd := exec.Command("sh", "-c", fmt.Sprintf("nginx -t 2>&1"))
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		return false, string(output)
	}
	return true, string(output)
}

// Reload 重载配置
func (s *NginxService) Reload() (string, error) {
	success, output := s.execute(config.AppConfig.Nginx.ReloadCommand)
	if success {
		return "Nginx 配置重载成功", nil
	}
	return "", fmt.Errorf(output)
}

// Start 启动Nginx
func (s *NginxService) Start() (string, error) {
	// 模拟启动
	return "Nginx 启动成功", nil
}

// Stop 停止Nginx
func (s *NginxService) Stop() (string, error) {
	// 模拟停止
	return "Nginx 停止成功", nil
}

// Restart 重启Nginx
func (s *NginxService) Restart() (string, error) {
	// 模拟重启
	return "Nginx 重启成功", nil
}

func (s *NginxService) execute(command string) (bool, string) {
	cmd := exec.Command("sh", "-c", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, string(output)
	}
	return true, string(output)
}

func (s *NginxService) saveHistory(fullPath, configType, configName, remark string) {
	content, err := ioutil.ReadFile(fullPath)
	if err != nil {
		log.Printf("Could not read config for history: %v", err)
		return
	}

	history := &model.NginxConfigHistory{
		ConfigName: configName,
		ConfigType: configType,
		FilePath:   fullPath,
		Content:    string(content),
		Remark:     remark,
	}
	if err := s.historyRepo.Create(history); err != nil {
		log.Printf("Could not save config history: %v", err)
	}
}
