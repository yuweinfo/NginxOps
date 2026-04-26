package service

import (
	"nginxops/internal/repository"
	"sort"
	"time"
)

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

// GetTrafficTrend 获取流量趋势数据
func (s *MetricsService) GetTrafficTrend(start, end time.Time, granularity string) []map[string]interface{} {
	collector := GetLogCollector()

	windowSeconds := parseGranularity(granularity)
	return collector.GetTrafficTrend(start, end, windowSeconds)
}

// GetResponseTrend 获取响应时间趋势
func (s *MetricsService) GetResponseTrend(start, end time.Time, granularity string) []map[string]interface{} {
	collector := GetLogCollector()

	windowSeconds := parseGranularity(granularity)
	return collector.GetResponseTimeTrend(start, end, windowSeconds)
}

// GetSlowRequestTrend 获取慢请求趋势
func (s *MetricsService) GetSlowRequestTrend(start, end time.Time, granularity string) []map[string]interface{} {
	collector := GetLogCollector()

	windowSeconds := parseGranularity(granularity)
	return collector.GetSlowRequestTrend(start, end, windowSeconds)
}

// GetMethodDistribution 获取请求方法分布
func (s *MetricsService) GetMethodDistribution(start, end time.Time) []map[string]interface{} {
	collector := GetLogCollector()
	return collector.GetMethodDistribution(start, end)
}

// GetStatusDistribution 获取状态码分布
func (s *MetricsService) GetStatusDistribution(start, end time.Time) []map[string]interface{} {
	collector := GetLogCollector()
	statusRank := collector.GetStatusRank(50)

	var total int64
	for _, item := range statusRank {
		total += item["count"].(int64)
	}

	result := make([]map[string]interface{}, 0, len(statusRank))
	for _, item := range statusRank {
		percent := float64(0)
		if total > 0 {
			percent = float64(item["count"].(int64)) / float64(total) * 100
		}
		result = append(result, map[string]interface{}{
			"name":    item["name"],
			"count":   item["count"],
			"percent": percent,
		})
	}
	return result
}

// GetErrorRateTrend 获取错误率趋势
func (s *MetricsService) GetErrorRateTrend(start, end time.Time, granularity string) []map[string]interface{} {
	collector := GetLogCollector()

	windowSeconds := parseGranularity(granularity)
	return collector.GetErrorRateTrend(start, end, windowSeconds)
}

// GetErrorPaths 获取错误路径 TOP
func (s *MetricsService) GetErrorPaths(start, end time.Time, limit int) []map[string]interface{} {
	logs, _, _ := s.accessLogRepo.FindPage(1, 10000, start, end, "")

	errorPaths := make(map[string]int64)
	var totalErrors int64
	for _, log := range logs {
		if log.Status >= 400 {
			errorPaths[log.Path]++
			totalErrors++
		}
	}

	type pathEntry struct {
		path  string
		count int64
	}
	entries := make([]pathEntry, 0, len(errorPaths))
	for path, count := range errorPaths {
		entries = append(entries, pathEntry{path: path, count: count})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].count > entries[j].count })

	if limit > len(entries) {
		limit = len(entries)
	}

	result := make([]map[string]interface{}, 0, limit)
	for i := 0; i < limit; i++ {
		percent := float64(0)
		if totalErrors > 0 {
			percent = float64(entries[i].count) / float64(totalErrors) * 100
		}
		result = append(result, map[string]interface{}{
			"path":    entries[i].path,
			"count":   entries[i].count,
			"percent": percent,
		})
	}
	return result
}

// GetClientAnalysis 获取客户端分析数据
func (s *MetricsService) GetClientAnalysis(start, end time.Time) map[string]interface{} {
	collector := GetLogCollector()

	deviceTypeRank := collector.GetDeviceTypeRank(20)
	browserRank := collector.GetBrowserRank(20)
	osRank := collector.GetOSRank(20)
	userAgentRank := collector.GetUserAgentRank(20)

	return map[string]interface{}{
		"deviceTypeRank": deviceTypeRank,
		"browserRank":    browserRank,
		"osRank":         osRank,
		"userAgentRank":  userAgentRank,
	}
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
