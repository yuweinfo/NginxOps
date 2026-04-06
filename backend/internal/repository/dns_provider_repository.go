package repository

import (
	"nginxops/internal/database"
	"nginxops/internal/model"
)

type DnsProviderRepository struct{}

func NewDnsProviderRepository() *DnsProviderRepository {
	return &DnsProviderRepository{}
}

func (r *DnsProviderRepository) FindAll() ([]model.DnsProvider, error) {
	var providers []model.DnsProvider
	if err := database.DB.Order("created_at DESC").Find(&providers).Error; err != nil {
		return nil, err
	}
	return providers, nil
}

func (r *DnsProviderRepository) FindByID(id uint) (*model.DnsProvider, error) {
	var provider model.DnsProvider
	if err := database.DB.First(&provider, id).Error; err != nil {
		return nil, err
	}
	return &provider, nil
}

func (r *DnsProviderRepository) Create(provider *model.DnsProvider) error {
	return database.DB.Create(provider).Error
}

func (r *DnsProviderRepository) Update(provider *model.DnsProvider) error {
	return database.DB.Save(provider).Error
}

func (r *DnsProviderRepository) Delete(id uint) error {
	return database.DB.Delete(&model.DnsProvider{}, id).Error
}

func (r *DnsProviderRepository) ClearDefault() error {
	return database.DB.Model(&model.DnsProvider{}).Where("is_default = ?", true).Update("is_default", false).Error
}

// FindDefault 查找默认DNS供应商
func (r *DnsProviderRepository) FindDefault() (*model.DnsProvider, error) {
	var provider model.DnsProvider
	if err := database.DB.Where("is_default = ?", true).First(&provider).Error; err != nil {
		return nil, err
	}
	return &provider, nil
}
