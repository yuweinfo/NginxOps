package repository

import (
	"nginxops/internal/database"
	"nginxops/internal/model"
)

type CertificateRequestRepository struct{}

func NewCertificateRequestRepository() *CertificateRequestRepository {
	return &CertificateRequestRepository{}
}

func (r *CertificateRequestRepository) Create(req *model.CertificateRequest) error {
	return database.DB.Create(req).Error
}

func (r *CertificateRequestRepository) Update(req *model.CertificateRequest) error {
	return database.DB.Save(req).Error
}

func (r *CertificateRequestRepository) FindByID(id uint) (*model.CertificateRequest, error) {
	var req model.CertificateRequest
	if err := database.DB.First(&req, id).Error; err != nil {
		return nil, err
	}
	return &req, nil
}
