package repository

import (
	"nginxops/internal/database"
	"nginxops/internal/model"
)

type SiteRepository struct{}

func NewSiteRepository() *SiteRepository {
	return &SiteRepository{}
}

func (r *SiteRepository) FindAll() ([]model.Site, error) {
	var sites []model.Site
	if err := database.DB.Order("created_at DESC").Find(&sites).Error; err != nil {
		return nil, err
	}
	return sites, nil
}

func (r *SiteRepository) FindByID(id uint) (*model.Site, error) {
	var site model.Site
	if err := database.DB.First(&site, id).Error; err != nil {
		return nil, err
	}
	return &site, nil
}

func (r *SiteRepository) FindByEnabled(enabled bool) ([]model.Site, error) {
	var sites []model.Site
	if err := database.DB.Where("enabled = ?", enabled).Find(&sites).Error; err != nil {
		return nil, err
	}
	return sites, nil
}

func (r *SiteRepository) FindPage(page, size int, keyword string) ([]model.Site, int64, error) {
	var sites []model.Site
	var total int64

	db := database.DB.Model(&model.Site{})
	if keyword != "" {
		db = db.Where("domain LIKE ? OR file_name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	if err := db.Order("created_at DESC").Offset(offset).Limit(size).Find(&sites).Error; err != nil {
		return nil, 0, err
	}

	return sites, total, nil
}

func (r *SiteRepository) Create(site *model.Site) error {
	return database.DB.Create(site).Error
}

func (r *SiteRepository) Update(site *model.Site) error {
	return database.DB.Save(site).Error
}

func (r *SiteRepository) Delete(id uint) error {
	return database.DB.Delete(&model.Site{}, id).Error
}
