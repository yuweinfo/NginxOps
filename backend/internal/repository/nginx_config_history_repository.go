package repository

import (
	"nginxops/internal/database"
	"nginxops/internal/model"
)

type NginxConfigHistoryRepository struct{}

func NewNginxConfigHistoryRepository() *NginxConfigHistoryRepository {
	return &NginxConfigHistoryRepository{}
}

func (r *NginxConfigHistoryRepository) FindByConfigName(configName string, limit int) ([]model.NginxConfigHistory, error) {
	var histories []model.NginxConfigHistory
	db := database.DB.Model(&model.NginxConfigHistory{})
	if configName != "" {
		db = db.Where("config_name = ?", configName)
	}
	if err := db.Order("created_at DESC").Limit(limit).Find(&histories).Error; err != nil {
		return nil, err
	}
	return histories, nil
}

func (r *NginxConfigHistoryRepository) Create(history *model.NginxConfigHistory) error {
	return database.DB.Create(history).Error
}
