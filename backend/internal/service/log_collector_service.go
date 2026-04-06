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
	statusCounts   map[string]*atomic.Int64 // 2xx, 3xx, 4xx, 5xx
	statusCountsMu sync.RWMutex

	// 小时趋势（内存聚合）
	hourlyCounts   map[int64]*atomic.Int64 // timestamp(hour) -> count
	hourlyCountsMu sync.RWMutex

	// IP 地理位置统计（内存聚合）
	ipLocationCounts   map[string]*IpLocationStat // ip -> stat
	ipLocationCountsMu sync.RWMutex

	// 地区排名（内存聚合）
	regionCounts   map[string]*atomic.Int64 // city -> count
	regionCountsMu sync.RWMutex

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
		logChan:          make(chan *model.AccessLog, channelSize),
		buffer:           make([]*model.AccessLog, 0, batchSize),
		qpsWindow:        make([]atomic.Int64, 60),
		bandwidthPerMin:  make([]atomic.Int64, 60),
		statusCounts:     make(map[string]*atomic.Int64),
		hourlyCounts:     make(map[int64]*atomic.Int64),
		ipLocationCounts: make(map[string]*IpLocationStat),
		regionCounts:     make(map[string]*atomic.Int64),
		stopCh:           make(chan struct{}),
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
}

func (s *LogCollectorService) incrementStatus(status int) {
	var key string
	switch {
	case status >= 200 && status < 300:
		key = "2xx"
	case status >= 300 && status < 400:
		key = "3xx"
	case status >= 400 && status < 500:
		key = "4xx"
	default:
		key = "5xx"
	}

	s.statusCountsMu.RLock()
	if counter, ok := s.statusCounts[key]; ok {
		counter.Add(1)
	}
	s.statusCountsMu.RUnlock()
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

		// 同时更新地区排名
		if geo.City != "" && geo.City != "Unknown" {
			s.doIncrementRegion(geo.City)
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

// GetIPLocations 获取 IP 地理位置分布（Top N）
func (s *LogCollectorService) GetIPLocations(limit int) []map[string]interface{} {
	s.ipLocationCountsMu.RLock()
	defer s.ipLocationCountsMu.RUnlock()

	// 转换为切片并排序
	type ipStat struct {
		IP       string
		Country  string
		Region   string
		City     string
		Lat      float64
		Lon      float64
		Requests int64
	}

	stats := make([]ipStat, 0, len(s.ipLocationCounts))
	for ip, stat := range s.ipLocationCounts {
		stats = append(stats, ipStat{
			IP:       ip,
			Country:  stat.Country,
			Region:   stat.Region,
			City:     stat.City,
			Lat:      stat.Lat,
			Lon:      stat.Lon,
			Requests: stat.Requests.Load(),
		})
	}

	// 简单排序（取 Top N）
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
		s := stats[i]
		if s.Lat != 0 && s.Lon != 0 {
			result = append(result, map[string]interface{}{
				"name":    s.City,
				"value":   []float64{s.Lon, s.Lat, float64(s.Requests)},
				"country": s.Country,
				"region":  s.Region,
			})
		}
	}
	return result
}

// GetRegionRank 获取地区排名（Top N）
func (s *LogCollectorService) GetRegionRank(limit int) []map[string]interface{} {
	s.regionCountsMu.RLock()
	defer s.regionCountsMu.RUnlock()

	type regionStat struct {
		City  string
		Count int64
	}

	stats := make([]regionStat, 0, len(s.regionCounts))
	for city, counter := range s.regionCounts {
		stats = append(stats, regionStat{
			City:  city,
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
			"city":    stats[i].City,
			"count":   stats[i].Count,
			"percent": percent,
		})
	}
	return result
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
}

// ========== 日志解析 ==========

var logPattern = regexp.MustCompile(`^(\S+) - (\S+) \[([^\]]+)\] "(\S+) ([^"]+) (\S+)" (\d+) (\d+) "([^"]*)" "([^"]*)" rt=([0-9.]+)`)

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
