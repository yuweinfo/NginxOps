package repository

import (
	"nginxops/internal/database"
	"nginxops/internal/model"
)

type UpstreamRepository struct{}

func NewUpstreamRepository() *UpstreamRepository {
	return &UpstreamRepository{}
}

func (r *UpstreamRepository) FindAll() ([]model.Upstream, error) {
	var upstreams []model.Upstream
	if err := database.DB.Order("created_at DESC").Find(&upstreams).Error; err != nil {
		return nil, err
	}
	return upstreams, nil
}

func (r *UpstreamRepository) FindByID(id uint) (*model.Upstream, error) {
	var upstream model.Upstream
	if err := database.DB.First(&upstream, id).Error; err != nil {
		return nil, err
	}
	return &upstream, nil
}

func (r *UpstreamRepository) FindPage(page, size int) ([]model.Upstream, int64, error) {
	var upstreams []model.Upstream
	var total int64

	db := database.DB.Model(&model.Upstream{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	if err := db.Order("created_at DESC").Offset(offset).Limit(size).Find(&upstreams).Error; err != nil {
		return nil, 0, err
	}

	return upstreams, total, nil
}

func (r *UpstreamRepository) Create(upstream *model.Upstream) error {
	return database.DB.Create(upstream).Error
}

func (r *UpstreamRepository) Update(upstream *model.Upstream) error {
	return database.DB.Save(upstream).Error
}

func (r *UpstreamRepository) Delete(id uint) error {
	return database.DB.Delete(&model.Upstream{}, id).Error
}
