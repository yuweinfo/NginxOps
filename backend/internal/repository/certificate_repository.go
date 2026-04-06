package repository

import (
	"nginxops/internal/database"
	"nginxops/internal/model"
)

type CertificateRepository struct{}

func NewCertificateRepository() *CertificateRepository {
	return &CertificateRepository{}
}

func (r *CertificateRepository) FindByID(id uint) (*model.Certificate, error) {
	var cert model.Certificate
	if err := database.DB.First(&cert, id).Error; err != nil {
		return nil, err
	}
	return &cert, nil
}

func (r *CertificateRepository) FindAll() ([]model.Certificate, error) {
	var certs []model.Certificate
	if err := database.DB.Order("created_at DESC").Find(&certs).Error; err != nil {
		return nil, err
	}
	return certs, nil
}

func (r *CertificateRepository) FindByStatus(status string) ([]model.Certificate, error) {
	var certs []model.Certificate
	if err := database.DB.Where("status = ?", status).Order("created_at DESC").Find(&certs).Error; err != nil {
		return nil, err
	}
	return certs, nil
}

func (r *CertificateRepository) FindByDomain(domain string) (*model.Certificate, error) {
	var cert model.Certificate
	if err := database.DB.Where("domain = ?", domain).First(&cert).Error; err != nil {
		return nil, err
	}
	return &cert, nil
}

func (r *CertificateRepository) FindPage(page, size int, status string) ([]model.Certificate, int64, error) {
	var certs []model.Certificate
	var total int64

	db := database.DB.Model(&model.Certificate{})
	if status != "" {
		db = db.Where("status = ?", status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	if err := db.Order("created_at DESC").Offset(offset).Limit(size).Find(&certs).Error; err != nil {
		return nil, 0, err
	}

	return certs, total, nil
}

func (r *CertificateRepository) Create(cert *model.Certificate) error {
	return database.DB.Create(cert).Error
}

func (r *CertificateRepository) Update(cert *model.Certificate) error {
	return database.DB.Save(cert).Error
}

func (r *CertificateRepository) Delete(id uint) error {
	return database.DB.Delete(&model.Certificate{}, id).Error
}
