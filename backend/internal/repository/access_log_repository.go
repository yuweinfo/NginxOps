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
