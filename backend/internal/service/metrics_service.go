package service

import (
	"nginxops/internal/repository"
	"sort"
	"time"
)

const metricsRankLimit = 20

type MetricsService struct {
	accessLogRepo *repository.AccessLogRepository
}

func NewMetricsService() *MetricsService {
	return &MetricsService{
		accessLogRepo: repository.NewAccessLogRepository(),
	}
}

// GetOverview 获取流量概览指标
func (s *MetricsService) GetOverview(start, end time.Time) map[string]interface{} {
	collector := GetLogCollector()

	totalRequests, totalBytes := collector.GetTotalRequestsAndBytes(start, end)
	peakQPS := collector.GetPeakQPS()
	avgRT := collector.GetAvgResponseTime(start, end)

	return map[string]interface{}{
		"totalRequests": totalRequests,
		"totalBytes":    totalBytes,
		"peakQPS":       peakQPS,
		"avgRT":         avgRT,
	}
}

// GetTrafficTrend 获取流量趋势数据（从数据库查询）
func (s *MetricsService) GetTrafficTrend(start, end time.Time, granularity string) []map[string]interface{} {
	windowSeconds := parseGranularity(granularity)
	points, err := s.accessLogRepo.GetTrafficTrend(start, end, windowSeconds)
	if err != nil {
		return nil
	}

	result := make([]map[string]interface{}, 0, len(points))
	for _, p := range points {
		result = append(result, map[string]interface{}{
			"time":     p.Time.Format("2006-01-02 15:04:05"),
			"requests": p.Requests,
			"bytes":    p.Bytes,
		})
	}
	return result
}

// GetResponseTrend 获取响应时间趋势（从数据库查询）
func (s *MetricsService) GetResponseTrend(start, end time.Time, granularity string) []map[string]interface{} {
	windowSeconds := parseGranularity(granularity)
	points, err := s.accessLogRepo.GetResponseTrend(start, end, windowSeconds)
	if err != nil {
		return nil
	}

	result := make([]map[string]interface{}, 0, len(points))
	for _, p := range points {
		result = append(result, map[string]interface{}{
			"time": p.Time.Format("2006-01-02 15:04:05"),
			"p50":  p.P50,
			"p90":  p.P90,
			"p99":  p.P99,
		})
	}
	return result
}

// GetSlowRequestTrend 获取慢请求趋势（从数据库查询）
func (s *MetricsService) GetSlowRequestTrend(start, end time.Time, granularity string) []map[string]interface{} {
	windowSeconds := parseGranularity(granularity)
	points, err := s.accessLogRepo.GetSlowRequestTrend(start, end, windowSeconds)
	if err != nil {
		return nil
	}

	result := make([]map[string]interface{}, 0, len(points))
	for _, p := range points {
		result = append(result, map[string]interface{}{
			"time":  p.Time.Format("2006-01-02 15:04:05"),
			"count": p.Count,
		})
	}
	return result
}

// GetMethodDistribution 获取请求方法分布
func (s *MetricsService) GetMethodDistribution(start, end time.Time) []map[string]interface{} {
	collector := GetLogCollector()
	return collector.GetMethodDistribution(start, end)
}

// GetStatusDistribution 获取状态码分布（从数据库查询，支持时间范围）
func (s *MetricsService) GetStatusDistribution(start, end time.Time) []map[string]interface{} {
	statusRank, err := s.accessLogRepo.GetStatusRank(start, end, 50)
	if err != nil {
		return nil
	}

	var total int64
	for _, item := range statusRank {
		total += item.Count
	}

	result := make([]map[string]interface{}, 0, len(statusRank))
	for _, item := range statusRank {
		percent := float64(0)
		if total > 0 {
			percent = float64(item.Count) / float64(total) * 100
		}
		result = append(result, map[string]interface{}{
			"name":    item.Name,
			"count":   item.Count,
			"percent": percent,
		})
	}
	return result
}

// GetErrorRateTrend 获取错误率趋势（从数据库查询）
func (s *MetricsService) GetErrorRateTrend(start, end time.Time, granularity string) []map[string]interface{} {
	windowSeconds := parseGranularity(granularity)
	points, err := s.accessLogRepo.GetErrorRateTrend(start, end, windowSeconds)
	if err != nil {
		return nil
	}

	result := make([]map[string]interface{}, 0, len(points))
	for _, p := range points {
		result = append(result, map[string]interface{}{
			"time":      p.Time.Format("2006-01-02 15:04:05"),
			"total":     p.Total,
			"errors":    p.Errors,
			"errorRate": p.ErrorRate,
		})
	}
	return result
}

// GetErrorPaths 获取错误路径 TOP（从数据库查询，使用 SQL 过滤而非全量加载）
func (s *MetricsService) GetErrorPaths(start, end time.Time, limit int) []map[string]interface{} {
	// 使用数据库查询直接过滤错误请求，避免加载全量数据到内存
	errorLogs, err := s.accessLogRepo.FindErrorPaths(start, end, limit)
	if err != nil {
		return nil
	}

	var totalErrors int64
	for _, item := range errorLogs {
		totalErrors += item.Count
	}

	result := make([]map[string]interface{}, 0, len(errorLogs))
	for _, item := range errorLogs {
		percent := float64(0)
		if totalErrors > 0 {
			percent = float64(item.Count) / float64(totalErrors) * 100
		}
		result = append(result, map[string]interface{}{
			"path":    item.Name,
			"count":   item.Count,
			"percent": percent,
		})
	}
	return result
}

// GetClientAnalysis 获取客户端分析数据（从数据库查询，支持时间范围）
func (s *MetricsService) GetClientAnalysis(start, end time.Time) map[string]interface{} {
	// 获取 user_agent 排行，在内存中解析分类
	uaRanks, err := s.accessLogRepo.GetUserAgentRank(start, end, metricsRankLimit*5)
	if err != nil {
		uaRanks = nil
	}

	browserCounts := make(map[string]int64)
	osCounts := make(map[string]int64)
	deviceTypeCounts := make(map[string]int64)

	collector := GetLogCollector()
	for _, item := range uaRanks {
		ua := item.Name
		if ua == "" {
			continue
		}
		count := item.Count
		browser := collector.ParseBrowser(ua)
		os := collector.ParseOS(ua)
		deviceType := collector.ParseDeviceType(ua)

		browserCounts[browser] += count
		osCounts[os] += count
		deviceTypeCounts[deviceType] += count
	}

	// user_agent 排行（原始值）
	uaResult := make([]map[string]interface{}, 0)
	if len(uaRanks) > metricsRankLimit {
		uaRanks = uaRanks[:metricsRankLimit]
	}
	var uaTotal int64
	for _, item := range uaRanks {
		uaTotal += item.Count
	}
	for _, item := range uaRanks {
		percent := float64(0)
		if uaTotal > 0 {
			percent = float64(item.Count) / float64(uaTotal) * 100
		}
		uaResult = append(uaResult, map[string]interface{}{
			"name":    item.Name,
			"count":   item.Count,
			"percent": percent,
		})
	}

	return map[string]interface{}{
		"deviceTypeRank": countsToMetricsRankResponse(deviceTypeCounts, metricsRankLimit),
		"browserRank":    countsToMetricsRankResponse(browserCounts, metricsRankLimit),
		"osRank":         countsToMetricsRankResponse(osCounts, metricsRankLimit),
		"userAgentRank":  uaResult,
	}
}

// countsToMetricsRankResponse 将 map[string]int64 转为排名格式
func countsToMetricsRankResponse(counts map[string]int64, limit int) []map[string]interface{} {
	type entry struct {
		name  string
		count int64
	}
	entries := make([]entry, 0, len(counts))
	for name, count := range counts {
		entries = append(entries, entry{name: name, count: count})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].count > entries[j].count
	})

	if limit > len(entries) {
		limit = len(entries)
	}

	var total int64
	for _, e := range entries {
		total += e.count
	}

	result := make([]map[string]interface{}, 0, limit)
	for i := 0; i < limit; i++ {
		percent := float64(0)
		if total > 0 {
			percent = float64(entries[i].count) / float64(total) * 100
		}
		result = append(result, map[string]interface{}{
			"name":    entries[i].name,
			"count":   entries[i].count,
			"percent": percent,
		})
	}
	return result
}

func parseGranularity(granularity string) int64 {
	switch granularity {
	case "1m":
		return 60
	case "5m":
		return 300
	case "1h":
		return 3600
	case "1d":
		return 86400
	default:
		return 300
	}
}
