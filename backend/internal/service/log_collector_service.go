package service

import (
	"log"
	"nginxops/internal/config"
	"nginxops/internal/model"
	"nginxops/internal/repository"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// LogCollectorService 日志收集服务
// 使用 channel + map 实现内存聚合统计，减少数据库读写
type LogCollectorService struct {
	logChan chan *model.AccessLog

	// 批量写入缓冲
	buffer   []*model.AccessLog
	bufferMu sync.Mutex

	// 实时统计（原子操作）
	todayPV       atomic.Int64
	todayUV       sync.Map // map[string]struct{}
	activeSites   atomic.Int32

	// 时间窗口统计（环形缓冲）
	qpsWindow     []atomic.Int64 // 每秒请求数，60个槽位
	qpsWindowMu   sync.Mutex
	qpsWindowPos  int

	// 状态码统计（内存聚合）
	statusCounts   map[string]*atomic.Int64 // 200, 301, 404, 500 etc.
	statusCountsMu sync.RWMutex

	// 小时趋势（内存聚合）
	hourlyCounts   map[int64]*atomic.Int64 // timestamp(hour) -> count
	hourlyCountsMu sync.RWMutex

	// IP 地理位置统计（内存聚合）
	ipLocationCounts   map[string]*IpLocationStat // ip -> stat
	ipLocationCountsMu sync.RWMutex

	// 地区排名（内存聚合，按国家）
	regionCounts   map[string]*atomic.Int64 // country -> count
	regionCountsMu sync.RWMutex

	// Host 排行
	hostCounts   map[string]*atomic.Int64
	hostCountsMu sync.RWMutex

	// Referer 排行
	refererCounts   map[string]*atomic.Int64
	refererCountsMu sync.RWMutex

	// URL Path 排行
	pathCounts   map[string]*atomic.Int64
	pathCountsMu sync.RWMutex

	// 资源类型排行
	resourceTypeCounts   map[string]*atomic.Int64
	resourceTypeCountsMu sync.RWMutex

	// 浏览器排行
	browserCounts   map[string]*atomic.Int64
	browserCountsMu sync.RWMutex

	// 设备类型排行
	deviceTypeCounts   map[string]*atomic.Int64
	deviceTypeCountsMu sync.RWMutex

	// 操作系统排行
	osCounts   map[string]*atomic.Int64
	osCountsMu sync.RWMutex

	// User-Agent 排行
	userAgentCounts   map[string]*atomic.Int64
	userAgentCountsMu sync.RWMutex

	// 带宽统计
	bandwidthPerMin []atomic.Int64 // 每分钟带宽，60个槽位
	bandwidthPos    atomic.Int32

	// 控制信号
	stopCh chan struct{}

	// 文件读取状态
	lastPosition int64
	lastInode    uint64
}

type IpLocationStat struct {
	Country  string
	Region   string
	City     string
	Lat      float64
	Lon      float64
	Requests atomic.Int64
}

const (
	channelSize   = 50000
	batchSize     = 1000
	flushInterval = 10 * time.Second
	dbWriteInterval = 30 * time.Second // 数据库写入间隔
)

var logCollectorInstance *LogCollectorService
var logCollectorOnce sync.Once

// GetLogCollector 获取单例
func GetLogCollector() *LogCollectorService {
	logCollectorOnce.Do(func() {
		logCollectorInstance = newLogCollector()
	})
	return logCollectorInstance
}

func newLogCollector() *LogCollectorService {
	s := &LogCollectorService{
		logChan:            make(chan *model.AccessLog, channelSize),
		buffer:             make([]*model.AccessLog, 0, batchSize),
		qpsWindow:          make([]atomic.Int64, 60),
		bandwidthPerMin:    make([]atomic.Int64, 60),
		statusCounts:       make(map[string]*atomic.Int64),
		hourlyCounts:       make(map[int64]*atomic.Int64),
		ipLocationCounts:   make(map[string]*IpLocationStat),
		regionCounts:       make(map[string]*atomic.Int64),
		hostCounts:         make(map[string]*atomic.Int64),
		refererCounts:      make(map[string]*atomic.Int64),
		pathCounts:         make(map[string]*atomic.Int64),
		resourceTypeCounts: make(map[string]*atomic.Int64),
		browserCounts:      make(map[string]*atomic.Int64),
		deviceTypeCounts:   make(map[string]*atomic.Int64),
		osCounts:           make(map[string]*atomic.Int64),
		userAgentCounts:    make(map[string]*atomic.Int64),
		stopCh:             make(chan struct{}),
	}

	// 初始化状态码计数器
	for _, code := range []string{"2xx", "3xx", "4xx", "5xx"} {
		s.statusCounts[code] = &atomic.Int64{}
	}

	return s
}

// Start 启动日志收集器
func (s *LogCollectorService) Start() {
	go s.readLogFile()
	go s.processLogs()
	go s.rotateStats()
	go s.periodicDBWrite()
	log.Println("Log collector started with in-memory aggregation")
}

// readLogFile 读取日志文件
func (s *LogCollectorService) readLogFile() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.readNewLogs()
		}
	}
}

func (s *LogCollectorService) readNewLogs() {
	accessLogPath := config.AppConfig.Nginx.AccessLog

	fileInfo, err := os.Stat(accessLogPath)
	if err != nil {
		return
	}

	// 检测文件轮转
	inode := getInode(accessLogPath)
	if inode != s.lastInode && s.lastInode != 0 {
		s.lastPosition = 0
	}
	s.lastInode = inode

	fileSize := fileInfo.Size()
	if fileSize <= s.lastPosition {
		return
	}

	file, err := os.Open(accessLogPath)
	if err != nil {
		return
	}
	defer file.Close()

	file.Seek(s.lastPosition, 0)
	buf := make([]byte, fileSize-s.lastPosition)
	n, _ := file.Read(buf)
	if n == 0 {
		return
	}

	s.lastPosition = fileSize

	// 解析并发送到 channel
	lines := strings.Split(string(buf[:n]), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		if parsed := s.parseLine(line); parsed != nil {
			select {
			case s.logChan <- parsed:
			default:
				// channel 满了，丢弃
			}
		}
	}
}

// processLogs 处理日志
func (s *LogCollectorService) processLogs() {
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			s.flushBuffer()
			return
		case entry := <-s.logChan:
			s.aggregateStats(entry)
			s.addToBuffer(entry)
		case <-ticker.C:
			s.flushBuffer()
		}
	}
}

// aggregateStats 内存聚合统计
func (s *LogCollectorService) aggregateStats(entry *model.AccessLog) {
	// 过滤健康检查请求
	if s.isHealthCheck(entry) {
		return
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// 判断是否今天
	if entry.TimeLocal.After(today) || entry.TimeLocal.Equal(today) {
		// PV
		s.todayPV.Add(1)

		// UV (记录 IP)
		s.todayUV.Store(entry.RemoteAddr, struct{}{})

		// 状态码统计
		s.incrementStatus(entry.Status)

		// 小时趋势
		hourTs := entry.TimeLocal.Unix() / 3600 * 3600
		s.incrementHourly(hourTs)

		// 带宽统计（当前分钟）
		minIdx := int(now.Second())
		if minIdx < 60 {
			s.bandwidthPerMin[minIdx].Add(entry.BodyBytes)
		}

		// QPS 窗口（当前秒）
		secIdx := int(now.Second())
		if secIdx < 60 {
			s.qpsWindowMu.Lock()
			s.qpsWindow[secIdx].Add(1)
			s.qpsWindowMu.Unlock()
		}
	}

	// IP 地理位置统计（异步查询）
	go s.updateIPLocation(entry.RemoteAddr)

	// Host 排行
	if entry.Host != "" {
		s.incrementHost(entry.Host)
	}

	// Referer 排行
	if entry.Referer != "" && entry.Referer != "-" {
		s.incrementReferer(entry.Referer)
	}

	// URL Path 排行
	if entry.Path != "" {
		s.incrementPath(entry.Path)
	}

	// 资源类型排行
	if entry.Path != "" {
		s.incrementResourceType(entry.Path)
	}

	// 浏览器排行
	if entry.UserAgent != "" && entry.UserAgent != "-" {
		s.incrementBrowser(entry.UserAgent)
	}

	// 设备类型排行
	if entry.UserAgent != "" && entry.UserAgent != "-" {
		s.incrementDeviceType(entry.UserAgent)
	}

	// 操作系统排行
	if entry.UserAgent != "" && entry.UserAgent != "-" {
		s.incrementOS(entry.UserAgent)
	}

	// User-Agent 排行
	if entry.UserAgent != "" && entry.UserAgent != "-" {
		s.incrementUserAgent(entry.UserAgent)
	}
}

// isHealthCheck 判断是否为健康检查请求
func (s *LogCollectorService) isHealthCheck(entry *model.AccessLog) bool {
	// 过滤本地健康检查请求
	if entry.RemoteAddr == "127.0.0.1" || entry.RemoteAddr == "::1" || entry.RemoteAddr == "localhost" {
		// 健康检查路径
		if entry.Path == "/api/health" || entry.Path == "/health" || entry.Path == "/" {
			return true
		}
	}
	return false
}

func (s *LogCollectorService) incrementStatus(status int) {
	key := strconv.Itoa(status)

	s.statusCountsMu.Lock()
	if counter, ok := s.statusCounts[key]; ok {
		counter.Add(1)
	} else {
		s.statusCounts[key] = &atomic.Int64{}
		s.statusCounts[key].Add(1)
	}
	s.statusCountsMu.Unlock()
}

func (s *LogCollectorService) incrementHourly(hourTs int64) {
	s.hourlyCountsMu.Lock()
	if counter, ok := s.hourlyCounts[hourTs]; ok {
		counter.Add(1)
	} else {
		s.hourlyCounts[hourTs] = &atomic.Int64{}
		s.hourlyCounts[hourTs].Add(1)
	}
	s.hourlyCountsMu.Unlock()
}

func (s *LogCollectorService) incrementHost(host string) {
	s.hostCountsMu.Lock()
	defer s.hostCountsMu.Unlock()
	if counter, ok := s.hostCounts[host]; ok {
		counter.Add(1)
	} else {
		s.hostCounts[host] = &atomic.Int64{}
		s.hostCounts[host].Add(1)
	}
}

func (s *LogCollectorService) incrementReferer(referer string) {
	s.refererCountsMu.Lock()
	defer s.refererCountsMu.Unlock()
	if counter, ok := s.refererCounts[referer]; ok {
		counter.Add(1)
	} else {
		s.refererCounts[referer] = &atomic.Int64{}
		s.refererCounts[referer].Add(1)
	}
}

func (s *LogCollectorService) incrementPath(path string) {
	s.pathCountsMu.Lock()
	defer s.pathCountsMu.Unlock()
	if counter, ok := s.pathCounts[path]; ok {
		counter.Add(1)
	} else {
		s.pathCounts[path] = &atomic.Int64{}
		s.pathCounts[path].Add(1)
	}
}

func (s *LogCollectorService) incrementResourceType(path string) {
	resourceType := s.getResourceType(path)
	s.resourceTypeCountsMu.Lock()
	defer s.resourceTypeCountsMu.Unlock()
	if counter, ok := s.resourceTypeCounts[resourceType]; ok {
		counter.Add(1)
	} else {
		s.resourceTypeCounts[resourceType] = &atomic.Int64{}
		s.resourceTypeCounts[resourceType].Add(1)
	}
}

func (s *LogCollectorService) getResourceType(path string) string {
	path = strings.ToLower(path)
	dotIdx := strings.LastIndex(path, ".")
	if dotIdx == -1 {
		return "Other"
	}
	ext := path[dotIdx:]
	switch {
	case ext == ".html", ext == ".htm":
		return "HTML"
	case ext == ".css":
		return "CSS"
	case ext == ".js", ext == ".mjs":
		return "JS"
	case ext == ".jpg", ext == ".jpeg", ext == ".png", ext == ".gif", ext == ".svg", ext == ".webp", ext == ".ico":
		return "Image"
	case ext == ".woff", ext == ".woff2", ext == ".ttf", ext == ".eot", ext == ".otf":
		return "Font"
	case ext == ".mp4", ext == ".webm", ext == ".avi", ext == ".mov", ext == ".flv":
		return "Video"
	case ext == ".mp3", ext == ".wav", ext == ".ogg", ext == ".flac":
		return "Audio"
	case ext == ".json", ext == ".xml":
		return "Data"
	case ext == ".pdf", ext == ".doc", ext == ".docx", ext == ".xls", ext == ".xlsx":
		return "Document"
	case ext == ".zip", ext == ".rar", ext == ".tar", ext == ".gz", ext == ".7z":
		return "Archive"
	default:
		return "Other"
	}
}

func (s *LogCollectorService) incrementBrowser(ua string) {
	browser := s.parseBrowser(ua)
	s.browserCountsMu.Lock()
	defer s.browserCountsMu.Unlock()
	if counter, ok := s.browserCounts[browser]; ok {
		counter.Add(1)
	} else {
		s.browserCounts[browser] = &atomic.Int64{}
		s.browserCounts[browser].Add(1)
	}
}

func (s *LogCollectorService) parseBrowser(ua string) string {
	uaLower := strings.ToLower(ua)
	switch {
	case strings.Contains(uaLower, "edg/") || strings.Contains(uaLower, "edge/"):
		return "Edge"
	case strings.Contains(uaLower, "chrome/") && !strings.Contains(uaLower, "edg/"):
		return "Chrome"
	case strings.Contains(uaLower, "firefox/"):
		return "Firefox"
	case strings.Contains(uaLower, "safari/") && !strings.Contains(uaLower, "chrome/"):
		return "Safari"
	case strings.Contains(uaLower, "opera/") || strings.Contains(uaLower, "opr/"):
		return "Opera"
	default:
		return "Other"
	}
}

func (s *LogCollectorService) incrementDeviceType(ua string) {
	deviceType := s.parseDeviceType(ua)
	s.deviceTypeCountsMu.Lock()
	defer s.deviceTypeCountsMu.Unlock()
	if counter, ok := s.deviceTypeCounts[deviceType]; ok {
		counter.Add(1)
	} else {
		s.deviceTypeCounts[deviceType] = &atomic.Int64{}
		s.deviceTypeCounts[deviceType].Add(1)
	}
}

func (s *LogCollectorService) parseDeviceType(ua string) string {
	uaLower := strings.ToLower(ua)
	if strings.Contains(uaLower, "mobile") || strings.Contains(uaLower, "android") ||
		strings.Contains(uaLower, "iphone") || strings.Contains(uaLower, "ipad") {
		return "Mobile"
	}
	return "Desktop"
}

func (s *LogCollectorService) incrementOS(ua string) {
	os := s.parseOS(ua)
	s.osCountsMu.Lock()
	defer s.osCountsMu.Unlock()
	if counter, ok := s.osCounts[os]; ok {
		counter.Add(1)
	} else {
		s.osCounts[os] = &atomic.Int64{}
		s.osCounts[os].Add(1)
	}
}

func (s *LogCollectorService) parseOS(ua string) string {
	uaLower := strings.ToLower(ua)
	switch {
	case strings.Contains(uaLower, "windows nt"):
		return "Windows"
	case strings.Contains(uaLower, "mac os") || strings.Contains(uaLower, "macos"):
		return "macOS"
	case strings.Contains(uaLower, "android"):
		return "Android"
	case strings.Contains(uaLower, "iphone") || strings.Contains(uaLower, "ipad"):
		return "iOS"
	case strings.Contains(uaLower, "linux"):
		return "Linux"
	default:
		return "Other"
	}
}

func (s *LogCollectorService) incrementUserAgent(ua string) {
	s.userAgentCountsMu.Lock()
	defer s.userAgentCountsMu.Unlock()
	if counter, ok := s.userAgentCounts[ua]; ok {
		counter.Add(1)
	} else {
		s.userAgentCounts[ua] = &atomic.Int64{}
		s.userAgentCounts[ua].Add(1)
	}
}

func (s *LogCollectorService) updateIPLocation(ip string) {
	// 先检查缓存
	s.ipLocationCountsMu.RLock()
	if _, exists := s.ipLocationCounts[ip]; exists {
		s.ipLocationCounts[ip].Requests.Add(1)
		s.ipLocationCountsMu.RUnlock()
		return
	}
	s.ipLocationCountsMu.RUnlock()

	// 获取地理位置
	geo := s.getGeoLocation(ip)
	if geo == nil {
		return
	}

	// 存储统计
	s.ipLocationCountsMu.Lock()
	if stat, exists := s.ipLocationCounts[ip]; exists {
		stat.Requests.Add(1)
	} else {
		stat := &IpLocationStat{
			Country: geo.Country,
			Region:  geo.Region,
			City:    geo.City,
			Lat:     geo.Lat,
			Lon:     geo.Lon,
		}
		stat.Requests.Add(1)
		s.ipLocationCounts[ip] = stat

		// 同时更新地区排名（按国家）
		if geo.Country != "" && geo.Country != "Unknown" {
			s.doIncrementRegion(geo.Country)
		}
	}
	s.ipLocationCountsMu.Unlock()
}

// getGeoLocation 获取 IP 地理位置（支持本地/内网 IP 模拟）
func (s *LogCollectorService) getGeoLocation(ip string) *IpLocationStat {
	// 本地/内网 IP 模拟数据（用于测试和展示）
	if isLocalOrPrivateIP(ip) {
		return getSimulatedLocation(ip)
	}

	// 查询真实地理位置
	geoSvc := NewGeoIpService()
	geo := geoSvc.GetGeo(ip)
	if geo == nil {
		return nil
	}

	return &IpLocationStat{
		Country: geo.Country,
		Region:  geo.Region,
		City:    geo.City,
		Lat:     geo.Lat,
		Lon:     geo.Lon,
	}
}

// isLocalOrPrivateIP 检查是否为本地或内网 IP
func isLocalOrPrivateIP(ip string) bool {
	return ip == "127.0.0.1" ||
		ip == "::1" ||
		ip == "localhost" ||
		strings.HasPrefix(ip, "192.168.") ||
		strings.HasPrefix(ip, "10.") ||
		strings.HasPrefix(ip, "172.16.") ||
		strings.HasPrefix(ip, "172.17.") ||
		strings.HasPrefix(ip, "172.18.") ||
		strings.HasPrefix(ip, "172.19.") ||
		strings.HasPrefix(ip, "172.20.") ||
		strings.HasPrefix(ip, "172.21.") ||
		strings.HasPrefix(ip, "172.22.") ||
		strings.HasPrefix(ip, "172.23.") ||
		strings.HasPrefix(ip, "172.24.") ||
		strings.HasPrefix(ip, "172.25.") ||
		strings.HasPrefix(ip, "172.26.") ||
		strings.HasPrefix(ip, "172.27.") ||
		strings.HasPrefix(ip, "172.28.") ||
		strings.HasPrefix(ip, "172.29.") ||
		strings.HasPrefix(ip, "172.30.") ||
		strings.HasPrefix(ip, "172.31.")
}

// getSimulatedLocation 为本地/内网 IP 返回模拟位置
func getSimulatedLocation(ip string) *IpLocationStat {
	// 根据不同 IP 返回不同模拟位置（均匀分布在全球主要城市）
	simulations := []IpLocationStat{
		{Country: "China", Region: "Beijing", City: "Beijing", Lat: 39.9042, Lon: 116.4074},
		{Country: "China", Region: "Shanghai", City: "Shanghai", Lat: 31.2304, Lon: 121.4737},
		{Country: "China", Region: "Guangdong", City: "Shenzhen", Lat: 22.5431, Lon: 114.0579},
		{Country: "China", Region: "Zhejiang", City: "Hangzhou", Lat: 30.2741, Lon: 120.1551},
		{Country: "United States", Region: "California", City: "Los Angeles", Lat: 34.0522, Lon: -118.2437},
		{Country: "United States", Region: "New York", City: "New York", Lat: 40.7128, Lon: -74.0060},
		{Country: "Japan", Region: "Tokyo", City: "Tokyo", Lat: 35.6762, Lon: 139.6503},
		{Country: "Singapore", Region: "Singapore", City: "Singapore", Lat: 1.3521, Lon: 103.8198},
		{Country: "Germany", Region: "Berlin", City: "Berlin", Lat: 52.5200, Lon: 13.4050},
		{Country: "United Kingdom", Region: "England", City: "London", Lat: 51.5074, Lon: -0.1278},
	}

	// 基于 IP 地址选择一个位置（相同 IP 始终返回相同位置）
	hash := 0
	for _, c := range ip {
		hash = hash*31 + int(c)
	}
	idx := hash % len(simulations)
	if idx < 0 {
		idx = -idx
	}

	return &simulations[idx]
}

func (s *LogCollectorService) doIncrementRegion(city string) {
	if counter, ok := s.regionCounts[city]; ok {
		counter.Add(1)
	} else {
		s.regionCounts[city] = &atomic.Int64{}
		s.regionCounts[city].Add(1)
	}
}

// rotateStats 定时清理过期统计
func (s *LogCollectorService) rotateStats() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			// 清理过期的小时数据（保留24小时）
			cutoff := time.Now().Add(-24 * time.Hour).Unix() / 3600 * 3600
			s.hourlyCountsMu.Lock()
			for ts := range s.hourlyCounts {
				if ts < cutoff {
					delete(s.hourlyCounts, ts)
				}
			}
			s.hourlyCountsMu.Unlock()

			// 重置 QPS 窗口（每分钟重置一次）
			s.qpsWindowMu.Lock()
			for i := range s.qpsWindow {
				s.qpsWindow[i].Store(0)
			}
			s.qpsWindowMu.Unlock()
		}
	}
}

// periodicDBWrite 定期写入数据库
func (s *LogCollectorService) periodicDBWrite() {
	ticker := time.NewTicker(dbWriteInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.flushBuffer()
		}
	}
}

func (s *LogCollectorService) addToBuffer(entry *model.AccessLog) {
	s.bufferMu.Lock()
	defer s.bufferMu.Unlock()

	s.buffer = append(s.buffer, entry)

	if len(s.buffer) >= batchSize {
		go s.flushBufferLocked()
	}
}

func (s *LogCollectorService) flushBuffer() {
	s.bufferMu.Lock()
	defer s.bufferMu.Unlock()
	s.flushBufferLocked()
}

func (s *LogCollectorService) flushBufferLocked() {
	if len(s.buffer) == 0 {
		return
	}

	logs := s.buffer
	s.buffer = make([]*model.AccessLog, 0, batchSize)

	go func(logs []*model.AccessLog) {
		if err := repository.BatchCreateAccessLogs(logs); err != nil {
			log.Printf("Failed to batch save access logs: %v", err)
		}
	}(logs)
}

// SetActiveSites 设置活跃站点数
func (s *LogCollectorService) SetActiveSites(count int) {
	s.activeSites.Store(int32(count))
}

// ========== 供 StatsService 读取的接口 ==========

// GetQPS 获取当前 QPS（最近1分钟平均值）
func (s *LogCollectorService) GetQPS() float64 {
	s.qpsWindowMu.Lock()
	defer s.qpsWindowMu.Unlock()

	var total int64
	for i := range s.qpsWindow {
		total += s.qpsWindow[i].Load()
	}
	return float64(total) / 60.0
}

// GetTodayPV 获取今日 PV
func (s *LogCollectorService) GetTodayPV() int64 {
	return s.todayPV.Load()
}

// GetTodayUV 获取今日 UV
func (s *LogCollectorService) GetTodayUV() int64 {
	var count int64
	s.todayUV.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

// GetActiveSites 获取活跃站点数
func (s *LogCollectorService) GetActiveSites() int {
	return int(s.activeSites.Load())
}

// GetBandwidth 获取当前带宽（最近1分钟）
func (s *LogCollectorService) GetBandwidth() float64 {
	var total int64
	for i := range s.bandwidthPerMin {
		total += s.bandwidthPerMin[i].Load()
	}
	return float64(total) / 1024 / 1024 // MB
}

// GetStatusDistribution 获取状态码分布
func (s *LogCollectorService) GetStatusDistribution() map[string]int64 {
	s.statusCountsMu.RLock()
	defer s.statusCountsMu.RUnlock()

	result := make(map[string]int64)
	for k, v := range s.statusCounts {
		result[k] = v.Load()
	}
	return result
}

// GetHourlyTrend 获取小时趋势（最近24小时）
func (s *LogCollectorService) GetHourlyTrend() []map[string]interface{} {
	s.hourlyCountsMu.RLock()
	defer s.hourlyCountsMu.RUnlock()

	result := make([]map[string]interface{}, 0)
	now := time.Now()
	for i := 23; i >= 0; i-- {
		t := now.Add(-time.Duration(i) * time.Hour)
		hourTs := t.Unix() / 3600 * 3600
		var count int64
		if counter, ok := s.hourlyCounts[hourTs]; ok {
			count = counter.Load()
		}
		result = append(result, map[string]interface{}{
			"hour":     t.Format("2006-01-02 15:04"),
			"requests": count,
		})
	}
	return result
}

// GetIPLocations 获取 IP 地理位置分布（按国家聚合，Top N）
func (s *LogCollectorService) GetIPLocations(limit int) []map[string]interface{} {
	s.ipLocationCountsMu.RLock()
	defer s.ipLocationCountsMu.RUnlock()

	// 按国家聚合
	type countryStat struct {
		Country  string
		Lat      float64
		Lon      float64
		Requests int64
	}

	countryMap := make(map[string]*countryStat)
	for _, stat := range s.ipLocationCounts {
		if c, ok := countryMap[stat.Country]; ok {
			c.Requests += stat.Requests.Load()
		} else {
			countryMap[stat.Country] = &countryStat{
				Country:  stat.Country,
				Lat:      stat.Lat,
				Lon:      stat.Lon,
				Requests: stat.Requests.Load(),
			}
		}
	}

	stats := make([]countryStat, 0, len(countryMap))
	for _, c := range countryMap {
		stats = append(stats, *c)
	}

	// 排序
	for i := 0; i < len(stats)-1; i++ {
		for j := i + 1; j < len(stats); j++ {
			if stats[j].Requests > stats[i].Requests {
				stats[i], stats[j] = stats[j], stats[i]
			}
		}
	}

	if limit > len(stats) {
		limit = len(stats)
	}

	result := make([]map[string]interface{}, 0, limit)
	for i := 0; i < limit; i++ {
		c := stats[i]
		if c.Lat != 0 && c.Lon != 0 {
			result = append(result, map[string]interface{}{
				"name":    c.Country,
				"value":   []float64{c.Lon, c.Lat, float64(c.Requests)},
				"country": c.Country,
			})
		}
	}
	return result
}

// GetRegionRank 获取国家排名（Top N）
func (s *LogCollectorService) GetRegionRank(limit int) []map[string]interface{} {
	s.regionCountsMu.RLock()
	defer s.regionCountsMu.RUnlock()

	type regionStat struct {
		Country string
		Count   int64
	}

	stats := make([]regionStat, 0, len(s.regionCounts))
	for country, counter := range s.regionCounts {
		stats = append(stats, regionStat{
			Country: country,
			Count:   counter.Load(),
		})
	}

	// 排序
	for i := 0; i < len(stats)-1; i++ {
		for j := i + 1; j < len(stats); j++ {
			if stats[j].Count > stats[i].Count {
				stats[i], stats[j] = stats[j], stats[i]
			}
		}
	}

	if limit > len(stats) {
		limit = len(stats)
	}

	var total int64
	for _, s := range stats {
		total += s.Count
	}

	result := make([]map[string]interface{}, 0, limit)
	for i := 0; i < limit; i++ {
		percent := float64(0)
		if total > 0 {
			percent = float64(stats[i].Count) / float64(total) * 100
		}
		result = append(result, map[string]interface{}{
			"country": stats[i].Country,
			"count":   stats[i].Count,
			"percent": percent,
		})
	}
	return result
}

// GetIPTopRank 获取IP访问排名（Top N）
func (s *LogCollectorService) GetIPTopRank(limit int) []map[string]interface{} {
	s.ipLocationCountsMu.RLock()
	defer s.ipLocationCountsMu.RUnlock()

	type ipStat struct {
		IP       string
		Country  string
		Requests int64
	}

	stats := make([]ipStat, 0, len(s.ipLocationCounts))
	for ip, stat := range s.ipLocationCounts {
		stats = append(stats, ipStat{
			IP:       ip,
			Country:  stat.Country,
			Requests: stat.Requests.Load(),
		})
	}

	// 排序（降序）
	for i := 0; i < len(stats)-1; i++ {
		for j := i + 1; j < len(stats); j++ {
			if stats[j].Requests > stats[i].Requests {
				stats[i], stats[j] = stats[j], stats[i]
			}
		}
	}

	if limit > len(stats) {
		limit = len(stats)
	}

	result := make([]map[string]interface{}, 0, limit)
	for i := 0; i < limit; i++ {
		result = append(result, map[string]interface{}{
			"ip":       stats[i].IP,
			"region":   stats[i].Country,
			"requests": stats[i].Requests,
		})
	}
	return result
}

// GetStatusRank 获取状态码排行（Top N）
func (s *LogCollectorService) GetStatusRank(limit int) []map[string]interface{} {
	s.statusCountsMu.RLock()
	defer s.statusCountsMu.RUnlock()

	type statusStat struct {
		Name  string
		Count int64
	}

	stats := make([]statusStat, 0, len(s.statusCounts))
	for name, counter := range s.statusCounts {
		stats = append(stats, statusStat{
			Name:  name,
			Count: counter.Load(),
		})
	}

	// 排序
	for i := 0; i < len(stats)-1; i++ {
		for j := i + 1; j < len(stats); j++ {
			if stats[j].Count > stats[i].Count {
				stats[i], stats[j] = stats[j], stats[i]
			}
		}
	}

	if limit > len(stats) {
		limit = len(stats)
	}

	var total int64
	for _, st := range stats {
		total += st.Count
	}

	result := make([]map[string]interface{}, 0, limit)
	for i := 0; i < limit; i++ {
		percent := float64(0)
		if total > 0 {
			percent = float64(stats[i].Count) / float64(total) * 100
		}
		result = append(result, map[string]interface{}{
			"name":    stats[i].Name,
			"count":   stats[i].Count,
			"percent": percent,
		})
	}
	return result
}

func (s *LogCollectorService) getRankData(counts map[string]*atomic.Int64, mu *sync.RWMutex, limit int) []map[string]interface{} {
	mu.RLock()
	defer mu.RUnlock()

	type rankStat struct {
		Name  string
		Count int64
	}

	stats := make([]rankStat, 0, len(counts))
	for name, counter := range counts {
		stats = append(stats, rankStat{
			Name:  name,
			Count: counter.Load(),
		})
	}

	// 排序
	for i := 0; i < len(stats)-1; i++ {
		for j := i + 1; j < len(stats); j++ {
			if stats[j].Count > stats[i].Count {
				stats[i], stats[j] = stats[j], stats[i]
			}
		}
	}

	if limit > len(stats) {
		limit = len(stats)
	}

	var total int64
	for _, st := range stats {
		total += st.Count
	}

	result := make([]map[string]interface{}, 0, limit)
	for i := 0; i < limit; i++ {
		percent := float64(0)
		if total > 0 {
			percent = float64(stats[i].Count) / float64(total) * 100
		}
		result = append(result, map[string]interface{}{
			"name":    stats[i].Name,
			"count":   stats[i].Count,
			"percent": percent,
		})
	}
	return result
}

// GetHostRank 获取 Host 排行（Top N）
func (s *LogCollectorService) GetHostRank(limit int) []map[string]interface{} {
	return s.getRankData(s.hostCounts, &s.hostCountsMu, limit)
}

// GetRefererRank 获取 Referer 排行（Top N）
func (s *LogCollectorService) GetRefererRank(limit int) []map[string]interface{} {
	return s.getRankData(s.refererCounts, &s.refererCountsMu, limit)
}

// GetPathRank 获取 URL Path 排行（Top N）
func (s *LogCollectorService) GetPathRank(limit int) []map[string]interface{} {
	return s.getRankData(s.pathCounts, &s.pathCountsMu, limit)
}

// GetResourceTypeRank 获取资源类型排行（Top N）
func (s *LogCollectorService) GetResourceTypeRank(limit int) []map[string]interface{} {
	return s.getRankData(s.resourceTypeCounts, &s.resourceTypeCountsMu, limit)
}

// GetBrowserRank 获取浏览器排行（Top N）
func (s *LogCollectorService) GetBrowserRank(limit int) []map[string]interface{} {
	return s.getRankData(s.browserCounts, &s.browserCountsMu, limit)
}

// GetDeviceTypeRank 获取设备类型排行（Top N）
func (s *LogCollectorService) GetDeviceTypeRank(limit int) []map[string]interface{} {
	return s.getRankData(s.deviceTypeCounts, &s.deviceTypeCountsMu, limit)
}

// GetOSRank 获取操作系统排行（Top N）
func (s *LogCollectorService) GetOSRank(limit int) []map[string]interface{} {
	return s.getRankData(s.osCounts, &s.osCountsMu, limit)
}

// GetUserAgentRank 获取 User-Agent 排行（Top N）
func (s *LogCollectorService) GetUserAgentRank(limit int) []map[string]interface{} {
	return s.getRankData(s.userAgentCounts, &s.userAgentCountsMu, limit)
}

// ResetDailyStats 每日重置统计（由定时任务调用）
func (s *LogCollectorService) ResetDailyStats() {
	s.todayPV.Store(0)
	s.todayUV = sync.Map{}

	s.statusCountsMu.Lock()
	for k := range s.statusCounts {
		s.statusCounts[k] = &atomic.Int64{}
	}
	s.statusCountsMu.Unlock()

	s.hostCountsMu.Lock()
	s.hostCounts = make(map[string]*atomic.Int64)
	s.hostCountsMu.Unlock()

	s.refererCountsMu.Lock()
	s.refererCounts = make(map[string]*atomic.Int64)
	s.refererCountsMu.Unlock()

	s.pathCountsMu.Lock()
	s.pathCounts = make(map[string]*atomic.Int64)
	s.pathCountsMu.Unlock()

	s.resourceTypeCountsMu.Lock()
	s.resourceTypeCounts = make(map[string]*atomic.Int64)
	s.resourceTypeCountsMu.Unlock()

	s.browserCountsMu.Lock()
	s.browserCounts = make(map[string]*atomic.Int64)
	s.browserCountsMu.Unlock()

	s.deviceTypeCountsMu.Lock()
	s.deviceTypeCounts = make(map[string]*atomic.Int64)
	s.deviceTypeCountsMu.Unlock()

	s.osCountsMu.Lock()
	s.osCounts = make(map[string]*atomic.Int64)
	s.osCountsMu.Unlock()

	s.userAgentCountsMu.Lock()
	s.userAgentCounts = make(map[string]*atomic.Int64)
	s.userAgentCountsMu.Unlock()

	s.regionCountsMu.Lock()
	s.regionCounts = make(map[string]*atomic.Int64)
	s.regionCountsMu.Unlock()
}

// ========== 日志解析 ==========

var logPattern = regexp.MustCompile(`^(\S+) - (\S+) \[([^\]]+)\] "(\S+) ([^"]+) (\S+)" (\d+) (\d+) "([^"]*)" "([^"]*)" rt=([0-9.]+)(?: host=(\S+))?`)

func (s *LogCollectorService) parseLine(line string) *model.AccessLog {
	matches := logPattern.FindStringSubmatch(line)
	if len(matches) < 12 {
		return nil
	}

	timeLocal, _ := time.Parse("02/Jan/2006:15:04:05 -0700", matches[3])
	status, _ := strconv.Atoi(matches[7])
	bodyBytes, _ := strconv.ParseInt(matches[8], 10, 64)
	rt, _ := strconv.ParseFloat(matches[11], 64)

	method := matches[4]
	path := matches[5]
	if strings.Contains(path, " ") {
		path = strings.Split(path, " ")[0]
	}
	if len(path) > 1024 {
		path = path[:1024]
	}

	host := ""
	if len(matches) >= 13 {
		host = matches[12]
	}

	return &model.AccessLog{
		RemoteAddr: matches[1],
		RemoteUser: nullIfDash(matches[2]),
		TimeLocal:  timeLocal,
		Request:    matches[4] + " " + matches[5] + " " + matches[6],
		Method:     method,
		Path:       path,
		Protocol:   matches[6],
		Status:     status,
		BodyBytes:  bodyBytes,
		Referer:    nullIfDash(matches[9]),
		UserAgent:  nullIfDash(matches[10]),
		RT:         rt,
		Host:       host,
	}
}

func nullIfDash(s string) string {
	if s == "-" {
		return ""
	}
	return s
}

func getInode(path string) uint64 {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return 0
	}
	if sys, ok := fileInfo.Sys().(interface{ Ino() uint64 }); ok {
		return sys.Ino()
	}
	return 0
}
