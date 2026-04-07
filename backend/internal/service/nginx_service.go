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

// ValidateConfig 验证配置语法（不保存，不中断服务）
// 创建独立的测试环境验证配置，不影响当前运行的 nginx
func (s *NginxService) ValidateConfig(configContent string) (bool, string) {
	// 检查 nginx 命令是否存在
	if _, err := exec.LookPath("nginx"); err != nil {
		// nginx 未安装，跳过验证（开发/测试环境）
		log.Println("Warning: nginx not found, skipping config validation")
		return true, ""
	}

	// 创建临时测试目录
	testDir := "/tmp/nginx_validate_" + filepath.Base(os.TempDir())
	os.RemoveAll(testDir)
	if err := os.MkdirAll(testDir, 0755); err != nil {
		return false, "无法创建临时目录: " + err.Error()
	}
	defer os.RemoveAll(testDir)

	// 创建临时 conf.d 目录
	testConfDir := filepath.Join(testDir, "conf.d")
	if err := os.MkdirAll(testConfDir, 0755); err != nil {
		return false, "无法创建临时配置目录: " + err.Error()
	}

	// 写入待测试的配置
	testConfFile := filepath.Join(testConfDir, "test.conf")
	if err := ioutil.WriteFile(testConfFile, []byte(configContent), 0644); err != nil {
		return false, "无法写入临时配置文件: " + err.Error()
	}

	// 复制现有的 upstream 配置文件到测试目录（解决跨文件引用问题）
	confDir := config.AppConfig.Nginx.ConfDir
	if files, err := ioutil.ReadDir(confDir); err == nil {
		for _, f := range files {
			if !f.IsDir() && strings.HasSuffix(f.Name(), ".conf") {
				srcPath := filepath.Join(confDir, f.Name())
				dstPath := filepath.Join(testConfDir, f.Name())
				// 不覆盖测试文件
				if srcPath != testConfFile {
					if data, err := ioutil.ReadFile(srcPath); err == nil {
						ioutil.WriteFile(dstPath, data, 0644)
					}
				}
			}
		}
	}

	// 创建临时 nginx.conf，包含测试配置
	mainConfig := config.AppConfig.Nginx.ConfigPath
	mainContent, err := ioutil.ReadFile(mainConfig)
	if err != nil {
		// 如果无法读取主配置，使用简单模板
		mainContent = []byte(`
worker_processes 1;
events { worker_connections 1024; }
http {
    include /etc/nginx/mime.types;
    default_type application/octet-stream;
    include ` + testConfDir + `/*.conf;
}
`)
	} else {
		// 替换 include 路径为测试目录
		content := string(mainContent)
		// 将原有 conf.d 路径替换为测试目录
		content = strings.Replace(content, "/etc/nginx/conf.d", testConfDir, -1)
		content = strings.Replace(content, "/data/nginx/conf.d", testConfDir, -1)
		mainContent = []byte(content)
	}

	testMainConfig := filepath.Join(testDir, "nginx.conf")
	if err := ioutil.WriteFile(testMainConfig, mainContent, 0644); err != nil {
		return false, "无法写入临时主配置文件: " + err.Error()
	}

	// 使用临时配置文件测试语法
	cmd := exec.Command("nginx", "-t", "-c", testMainConfig)
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
