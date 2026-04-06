package repository

import (
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
