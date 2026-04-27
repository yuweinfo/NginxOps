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
