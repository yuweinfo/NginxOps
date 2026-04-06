package repository

import (
	"nginxops/internal/database"
	"nginxops/internal/model"
)

type IpGeoCacheRepository struct{}

func NewIpGeoCacheRepository() *IpGeoCacheRepository {
	return &IpGeoCacheRepository{}
}

func (r *IpGeoCacheRepository) FindByIP(ip string) (*model.IpGeoCache, error) {
	var cache model.IpGeoCache
	if err := database.DB.First(&cache, "ip = ?", ip).Error; err != nil {
		return nil, err
	}
	return &cache, nil
}

func (r *IpGeoCacheRepository) Create(cache *model.IpGeoCache) error {
	return database.DB.Create(cache).Error
}

func (r *IpGeoCacheRepository) Update(cache *model.IpGeoCache) error {
	return database.DB.Save(cache).Error
}
