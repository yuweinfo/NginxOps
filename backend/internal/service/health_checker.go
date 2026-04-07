package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"nginxops/internal/model"
	"nginxops/internal/repository"
)

// HealthChecker 健康检查服务
type HealthChecker struct {
	upstreamRepo *repository.UpstreamRepository
	client       *http.Client
	stopChan     chan struct{}
	running      bool
	mu           sync.Mutex
}

// HealthCheckResult 健康检查结果
type HealthCheckResult struct {
	UpstreamID   uint   `json:"upstreamId"`
	UpstreamName string `json:"upstreamName"`
	ServerHost   string `json:"serverHost"`
	ServerPort   int    `json:"serverPort"`
	Healthy      bool   `json:"healthy"`
	ResponseTime int64  `json:"responseTime"` // 毫秒
	Error        string `json:"error,omitempty"`
	CheckedAt    string `json:"checkedAt"`
}

// ServerHealthStatus 服务器健康状态
type ServerHealthStatus struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Healthy      bool   `json:"healthy"`
	ResponseTime int64  `json:"responseTime"`
	LastCheck    string `json:"lastCheck"`
	Error        string `json:"error,omitempty"`
}

var (
	healthChecker     *HealthChecker
	healthCheckerOnce sync.Once
)

// GetHealthChecker 获取健康检查服务单例
func GetHealthChecker() *HealthChecker {
	healthCheckerOnce.Do(func() {
		healthChecker = &HealthChecker{
			upstreamRepo: repository.NewUpstreamRepository(),
			client: &http.Client{
				Timeout: 10 * time.Second,
			},
			stopChan: make(chan struct{}),
		}
	})
	return healthChecker
}

// Start 启动健康检查
func (h *HealthChecker) Start(intervalSeconds int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.running {
		return
	}

	h.running = true
	go h.runChecks(intervalSeconds)
}

// Stop 停止健康检查
func (h *HealthChecker) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.running {
		return
	}

	close(h.stopChan)
	h.running = false
	// 重新创建 stopChan 以便后续可以重新启动
	h.stopChan = make(chan struct{})
}

// runChecks 执行定期检查
func (h *HealthChecker) runChecks(intervalSeconds int) {
	ticker := time.NewTicker(time.Duration(intervalSeconds) * time.Second)
	defer ticker.Stop()

	// 首次立即执行一次
	h.checkAll()

	for {
		select {
		case <-ticker.C:
			h.checkAll()
		case <-h.stopChan:
			return
		}
	}
}

// checkAll 检查所有 upstream
func (h *HealthChecker) checkAll() {
	upstreams, err := h.upstreamRepo.FindAll()
	if err != nil {
		return
	}

	for _, upstream := range upstreams {
		if !upstream.HealthCheck {
			continue
		}
		h.checkUpstream(&upstream)
	}
}

// checkUpstream 检查单个 upstream 的所有服务器
func (h *HealthChecker) checkUpstream(upstream *model.Upstream) {
	var servers []model.UpstreamServer
	if err := json.Unmarshal([]byte(upstream.Servers), &servers); err != nil {
		return
	}

	hasChanges := false
	checkPath := upstream.CheckPath
	if checkPath == "" {
		checkPath = "/health"
	}
	timeout := upstream.CheckTimeout
	if timeout == 0 {
		timeout = 3
	}

	for i, server := range servers {
		if server.Backup {
			// 备用服务器不主动检查
			continue
		}

		result := h.checkServer(server.Host, server.Port, checkPath, timeout)
		newStatus := "up"
		if !result.Healthy {
			newStatus = "down"
		}

		if server.Status != newStatus {
			servers[i].Status = newStatus
			hasChanges = true
		}
	}

	if hasChanges {
		updatedServers, _ := json.Marshal(servers)
		upstream.Servers = string(updatedServers)
		h.upstreamRepo.Update(upstream)
	}
}

// checkServer 检查单个服务器
func (h *HealthChecker) checkServer(host string, port int, path string, timeout int) HealthCheckResult {
	result := HealthCheckResult{
		ServerHost: host,
		ServerPort: port,
		CheckedAt:  time.Now().Format(time.RFC3339),
	}

	url := fmt.Sprintf("http://%s:%d%s", host, port, path)

	start := time.Now()
	resp, err := h.client.Get(url)
	result.ResponseTime = time.Since(start).Milliseconds()

	if err != nil {
		result.Healthy = false
		result.Error = err.Error()
		return result
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 500 {
		result.Healthy = true
	} else {
		result.Healthy = false
		result.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
	}

	return result
}

// CheckUpstreamNow 立即检查指定 upstream
func (h *HealthChecker) CheckUpstreamNow(upstreamID uint) ([]HealthCheckResult, error) {
	upstream, err := h.upstreamRepo.FindByID(upstreamID)
	if err != nil {
		return nil, fmt.Errorf("upstream 不存在")
	}

	var servers []model.UpstreamServer
	if err := json.Unmarshal([]byte(upstream.Servers), &servers); err != nil {
		return nil, fmt.Errorf("服务器配置解析失败")
	}

	checkPath := upstream.CheckPath
	if checkPath == "" {
		checkPath = "/health"
	}
	timeout := upstream.CheckTimeout
	if timeout == 0 {
		timeout = 3
	}

	var results []HealthCheckResult
	hasChanges := false

	for i, server := range servers {
		result := h.checkServer(server.Host, server.Port, checkPath, timeout)
		result.UpstreamID = upstream.ID
		result.UpstreamName = upstream.Name

		newStatus := "up"
		if !result.Healthy {
			newStatus = "down"
		}

		if server.Status != newStatus {
			servers[i].Status = newStatus
			hasChanges = true
		}

		results = append(results, result)
	}

	if hasChanges {
		updatedServers, _ := json.Marshal(servers)
		upstream.Servers = string(updatedServers)
		h.upstreamRepo.Update(upstream)
	}

	return results, nil
}

// GetServersHealth 获取服务器健康状态
func (h *HealthChecker) GetServersHealth(upstreamID uint) ([]ServerHealthStatus, error) {
	upstream, err := h.upstreamRepo.FindByID(upstreamID)
	if err != nil {
		return nil, fmt.Errorf("upstream 不存在")
	}

	var servers []model.UpstreamServer
	if err := json.Unmarshal([]byte(upstream.Servers), &servers); err != nil {
		return nil, fmt.Errorf("服务器配置解析失败")
	}

	var statuses []ServerHealthStatus
	for _, server := range servers {
		statuses = append(statuses, ServerHealthStatus{
			Host:    server.Host,
			Port:    server.Port,
			Healthy: server.Status == "up",
		})
	}

	return statuses, nil
}
