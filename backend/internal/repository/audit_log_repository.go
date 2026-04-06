package repository

import (
	"nginxops/internal/database"
	"nginxops/internal/model"
	"time"
)

type AuditLogRepository struct{}

func NewAuditLogRepository() *AuditLogRepository {
	return &AuditLogRepository{}
}

func (r *AuditLogRepository) FindPage(page, size int, module, action string, userID uint, status string, startTime, endTime *time.Time) ([]model.AuditLog, int64, error) {
	var logs []model.AuditLog
	var total int64

	db := database.DB.Model(&model.AuditLog{})
	if module != "" {
		db = db.Where("module = ?", module)
	}
	if action != "" {
		db = db.Where("action = ?", action)
	}
	if userID > 0 {
		db = db.Where("user_id = ?", userID)
	}
	if status != "" {
		db = db.Where("status = ?", status)
	}
	if startTime != nil {
		db = db.Where("created_at >= ?", startTime)
	}
	if endTime != nil {
		db = db.Where("created_at <= ?", endTime)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	if err := db.Order("created_at DESC").Offset(offset).Limit(size).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (r *AuditLogRepository) FindByID(id uint) (*model.AuditLog, error) {
	var log model.AuditLog
	if err := database.DB.First(&log, id).Error; err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *AuditLogRepository) Create(log *model.AuditLog) error {
	return database.DB.Create(log).Error
}
