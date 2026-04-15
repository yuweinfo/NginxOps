package repository

import (
	"encoding/json"
	"nginxops/internal/database"
	"nginxops/internal/model"
)

type CountryCIDRRepository struct{}

func NewCountryCIDRRepository() *CountryCIDRRepository {
	return &CountryCIDRRepository{}
}

func (r *CountryCIDRRepository) FindByCountryCode(code string) ([]string, error) {
	var record model.CountryCIDR
	if err := database.DB.Where("country_code = ?", code).First(&record).Error; err != nil {
		return nil, err
	}
	var cidrs []string
	if err := json.Unmarshal([]byte(record.CIDRs), &cidrs); err != nil {
		return nil, err
	}
	return cidrs, nil
}

func (r *CountryCIDRRepository) Save(code string, cidrs []string) error {
	data, err := json.Marshal(cidrs)
	if err != nil {
		return err
	}
	var record model.CountryCIDR
	err = database.DB.Where("country_code = ?", code).First(&record).Error
	if err != nil {
		// 创建
		record = model.CountryCIDR{
			CountryCode: code,
			CIDRs:       string(data),
		}
		return database.DB.Create(&record).Error
	}
	record.CIDRs = string(data)
	return database.DB.Save(&record).Error
}
