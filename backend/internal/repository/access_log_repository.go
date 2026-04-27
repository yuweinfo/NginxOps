package repository

import (
	"fmt"
	"nginxops/internal/database"
	"nginxops/internal/model"
	"time"
)

type AccessLogRepository struct{}

func NewAccessLogRepository() *AccessLogRepository {
	return &AccessLogRepository{}
}

func (r *AccessLogRepository) FindPage(page, size int, start, end time.Time, ip string) ([]model.AccessLog, int64, error) {
	var logs = make([]model.AccessLog, 0)
	var total int64

	db := database.DB.Model(&model.AccessLog{})
	if !start.IsZero() {
		db = db.Where("time_local >= ?", start)
	}
	if !end.IsZero() {
		db = db.Where("time_local <= ?", end)
	}
	if ip != "" {
		db = db.Where("remote_addr = ?", ip)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	if err := db.Order("time_local DESC").Offset(offset).Limit(size).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (r *AccessLogRepository) CountByStatus(start, end time.Time) (map[string]int64, error) {
	type StatusCount struct {
		Status string
		Count  int64
	}
	var results []StatusCount

	err := database.DB.Model(&model.AccessLog{}).
		Select("CASE WHEN status >= 200 AND status < 300 THEN '2xx' WHEN status >= 300 AND status < 400 THEN '3xx' WHEN status >= 400 AND status < 500 THEN '4xx' ELSE '5xx' END as status, COUNT(*) as count").
		Where("time_local >= ? AND time_local <= ?", start, end).
		Group("status").
		Find(&results).Error

	counts := make(map[string]int64)
	for _, r := range results {
		counts[r.Status] = r.Count
	}
	return counts, err
}

func (r *AccessLogRepository) CountPV(start, end time.Time) (int64, error) {
	var count int64
	err := database.DB.Model(&model.AccessLog{}).
		Where("time_local >= ? AND time_local <= ?", start, end).
		Count(&count).Error
	return count, err
}

func (r *AccessLogRepository) CountUV(start, end time.Time) (int64, error) {
	var count int64
	err := database.DB.Model(&model.AccessLog{}).
		Where("time_local >= ? AND time_local <= ?", start, end).
		Distinct("remote_addr").
		Count(&count).Error
	return count, err
}

// HourlyCount 按小时统计请求数
type HourlyCount struct {
	Hour     string
	Requests int64
}

func (r *AccessLogRepository) CountByHour(start, end time.Time) ([]HourlyCount, error) {
	var results []HourlyCount
	err := database.DB.Model(&model.AccessLog{}).
		Select("TO_CHAR(time_local, 'YYYY-MM-DD HH24:00') as hour, COUNT(*) as requests").
		Where("time_local >= ? AND time_local <= ?", start, end).
		Group("hour").
		Order("hour").
		Find(&results).Error
	return results, err
}

// IpLocationCount IP地理位置统计
type IpLocationCount struct {
	IP       string
	Country  string
	Region   string
	City     string
	Lat      float64
	Lon      float64
	Requests int64
}

func (r *AccessLogRepository) CountByLocation(start, end time.Time, limit int) ([]IpLocationCount, error) {
	var results []IpLocationCount
	err := database.DB.Table("access_log a").
		Select("a.remote_addr as ip, g.country, g.region, g.city, g.lat, g.lon, COUNT(*) as requests").
		Joins("LEFT JOIN ip_geo_cache g ON a.remote_addr = g.ip").
		Where("a.time_local >= ? AND a.time_local <= ?", start, end).
		Group("a.remote_addr, g.country, g.region, g.city, g.lat, g.lon").
		Order("requests DESC").
		Limit(limit).
		Find(&results).Error
	return results, err
}

// RegionCount 地区访问统计
type RegionCount struct {
	City   string
	Count  int64
}

func (r *AccessLogRepository) CountByRegion(start, end time.Time, limit int) ([]RegionCount, error) {
	var results []RegionCount
	err := database.DB.Table("access_log a").
		Select("COALESCE(g.city, 'Unknown') as city, COUNT(*) as count").
		Joins("LEFT JOIN ip_geo_cache g ON a.remote_addr = g.ip").
		Where("a.time_local >= ? AND a.time_local <= ?", start, end).
		Group("g.city").
		Order("count DESC").
		Limit(limit).
		Find(&results).Error
	return results, err
}

// SumBandwidth 统计带宽总量
func (r *AccessLogRepository) SumBandwidth(start, end time.Time) (int64, error) {
	var total int64
	err := database.DB.Model(&model.AccessLog{}).
		Where("time_local >= ? AND time_local <= ?", start, end).
		Select("COALESCE(SUM(body_bytes), 0)").
		Scan(&total).Error
	return total, err
}

// BatchCreateAccessLogs 批量创建访问日志
func BatchCreateAccessLogs(logs []*model.AccessLog) error {
	if len(logs) == 0 {
		return nil
	}
	return database.DB.CreateInBatches(logs, 100).Error
}

// TrafficTrendPoint 流量趋势数据点
type TrafficTrendPoint struct {
	Time     time.Time `json:"time"`
	Requests int64     `json:"requests"`
	Bytes    int64     `json:"bytes"`
}

// ResponseTrendPoint 响应时间趋势数据点
type ResponseTrendPoint struct {
	Time time.Time `json:"time"`
	P50  float64   `json:"p50"`
	P90  float64   `json:"p90"`
	P99  float64   `json:"p99"`
}

// SlowRequestTrendPoint 慢请求趋势数据点
type SlowRequestTrendPoint struct {
	Time  time.Time `json:"time"`
	Count int64     `json:"count"`
}

// ErrorRateTrendPoint 错误率趋势数据点
type ErrorRateTrendPoint struct {
	Time      time.Time `json:"time"`
	Total     int64     `json:"total"`
	Errors    int64     `json:"errors"`
	ErrorRate float64   `json:"errorRate"`
}

func getTimeBucketExpr(windowSeconds int64) string {
	switch windowSeconds {
	case 60:
		return "DATE_TRUNC('minute', time_local)"
	case 300:
		return "TO_TIMESTAMP(FLOOR(EXTRACT(EPOCH FROM time_local) / 300) * 300)"
	case 3600:
		return "DATE_TRUNC('hour', time_local)"
	case 86400:
		return "DATE_TRUNC('day', time_local)"
	default:
		return "TO_TIMESTAMP(FLOOR(EXTRACT(EPOCH FROM time_local) / " + fmt.Sprintf("%d", windowSeconds) + ") * " + fmt.Sprintf("%d", windowSeconds) + ")"
	}
}

// GetTrafficTrend 获取流量趋势（从数据库）
func (r *AccessLogRepository) GetTrafficTrend(start, end time.Time, windowSeconds int64) ([]TrafficTrendPoint, error) {
	var results []TrafficTrendPoint
	bucketExpr := getTimeBucketExpr(windowSeconds)
	err := database.DB.Model(&model.AccessLog{}).
		Select(fmt.Sprintf("%s as time, COUNT(*) as requests, COALESCE(SUM(body_bytes), 0) as bytes", bucketExpr)).
		Where("time_local >= ? AND time_local <= ?", start, end).
		Group("time").
		Order("time").
		Find(&results).Error
	return results, err
}

// GetResponseTrend 获取响应时间趋势（P50/P90/P99，从数据库）
func (r *AccessLogRepository) GetResponseTrend(start, end time.Time, windowSeconds int64) ([]ResponseTrendPoint, error) {
	var results []ResponseTrendPoint
	bucketExpr := getTimeBucketExpr(windowSeconds)
	
	err := database.DB.Model(&model.AccessLog{}).
		Select(fmt.Sprintf("%s as time, "+
			"PERCENTILE_CONT(0.50) WITHIN GROUP (ORDER BY rt) as p50, "+
			"PERCENTILE_CONT(0.90) WITHIN GROUP (ORDER BY rt) as p90, "+
			"PERCENTILE_CONT(0.99) WITHIN GROUP (ORDER BY rt) as p99", bucketExpr)).
		Where("time_local >= ? AND time_local <= ?", start, end).
		Group("time").
		Order("time").
		Find(&results).Error
	return results, err
}

// GetSlowRequestTrend 获取慢请求趋势（RT > 1s，从数据库）
func (r *AccessLogRepository) GetSlowRequestTrend(start, end time.Time, windowSeconds int64) ([]SlowRequestTrendPoint, error) {
	var results []SlowRequestTrendPoint
	bucketExpr := getTimeBucketExpr(windowSeconds)
	
	err := database.DB.Model(&model.AccessLog{}).
		Select(fmt.Sprintf("%s as time, COUNT(*) as count", bucketExpr)).
		Where("time_local >= ? AND time_local <= ? AND rt > 1", start, end).
		Group("time").
		Order("time").
		Find(&results).Error
	return results, err
}

// GetErrorRateTrend 获取错误率趋势（从数据库）
func (r *AccessLogRepository) GetErrorRateTrend(start, end time.Time, windowSeconds int64) ([]ErrorRateTrendPoint, error) {
	type rawTrend struct {
		Time   time.Time `gorm:"column:time"`
		Total  int64
		Errors int64
	}
	var rawResults []rawTrend
	bucketExpr := getTimeBucketExpr(windowSeconds)
	
	err := database.DB.Model(&model.AccessLog{}).
		Select(fmt.Sprintf("%s as time, COUNT(*) as total, COUNT(*) FILTER (WHERE status >= 400) as errors", bucketExpr)).
		Where("time_local >= ? AND time_local <= ?", start, end).
		Group("time").
		Order("time").
		Find(&rawResults).Error
	if err != nil {
		return nil, err
	}

	results := make([]ErrorRateTrendPoint, 0, len(rawResults))
	for _, r := range rawResults {
		errorRate := float64(0)
		if r.Total > 0 {
			errorRate = float64(r.Errors) / float64(r.Total) * 100
		}
		results = append(results, ErrorRateTrendPoint{
			Time:      r.Time,
			Total:     r.Total,
			Errors:    r.Errors,
			ErrorRate: errorRate,
		})
	}
	return results, nil
}
