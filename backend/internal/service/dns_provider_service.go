package service

import (
	"fmt"
	"nginxops/internal/model"
	"nginxops/internal/repository"
)

type DnsProviderService struct {
	repo *repository.DnsProviderRepository
}

func NewDnsProviderService() *DnsProviderService {
	return &DnsProviderService{
		repo: repository.NewDnsProviderRepository(),
	}
}

func (s *DnsProviderService) ListAll() ([]model.DnsProvider, error) {
	return s.repo.FindAll()
}

func (s *DnsProviderService) GetByID(id uint) (*model.DnsProvider, error) {
	return s.repo.FindByID(id)
}

func (s *DnsProviderService) Create(provider *model.DnsProvider) (*model.DnsProvider, error) {
	if provider.IsDefault {
		if err := s.repo.ClearDefault(); err != nil {
			return nil, err
		}
	}
	if err := s.repo.Create(provider); err != nil {
		return nil, err
	}
	return provider, nil
}

func (s *DnsProviderService) Update(id uint, provider *model.DnsProvider) (*model.DnsProvider, error) {
	existing, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("DNS 服务商配置不存在")
	}

	existing.Name = provider.Name
	existing.ProviderType = provider.ProviderType
	existing.AccessKeyID = provider.AccessKeyID
	if provider.AccessKeySecret != "" {
		existing.AccessKeySecret = provider.AccessKeySecret
	}
	existing.IsDefault = provider.IsDefault

	if existing.IsDefault {
		if err := s.repo.ClearDefault(); err != nil {
			return nil, err
		}
	}

	if err := s.repo.Update(existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *DnsProviderService) Delete(id uint) error {
	return s.repo.Delete(id)
}

func (s *DnsProviderService) SetDefault(id uint) error {
	provider, err := s.repo.FindByID(id)
	if err != nil {
		return fmt.Errorf("DNS 服务商配置不存在")
	}

	if err := s.repo.ClearDefault(); err != nil {
		return err
	}

	provider.IsDefault = true
	return s.repo.Update(provider)
}
