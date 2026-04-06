package service

import (
	"nginxops/internal/model"
	"nginxops/internal/repository"
	"time"
)

type AuditService struct {
	repo *repository.AuditLogRepository
}

func NewAuditService() *AuditService {
	return &AuditService{
		repo: repository.NewAuditLogRepository(),
	}
}

type AuditQueryResult struct {
	List  []model.AuditLog `json:"list"`
	Total int64            `json:"total"`
	Page  int              `json:"page"`
	Size  int              `json:"size"`
}

func (s *AuditService) Query(page, size int, module, action string, userID uint, status string, startTime, endTime *time.Time) (*AuditQueryResult, error) {
	logs, total, err := s.repo.FindPage(page, size, module, action, userID, status, startTime, endTime)
	if err != nil {
		return nil, err
	}
	return &AuditQueryResult{
		List:  logs,
		Total: total,
		Page:  page,
		Size:  size,
	}, nil
}

func (s *AuditService) GetByID(id uint) (*model.AuditLog, error) {
	return s.repo.FindByID(id)
}

func (s *AuditService) Log(userID uint, username, action, module, targetType string, targetID uint, targetName, detail, ip, userAgent, status, errorMsg string) error {
	log := &model.AuditLog{
		UserID:     userID,
		Username:   username,
		Action:     action,
		Module:     module,
		TargetType: targetType,
		TargetID:   targetID,
		TargetName: targetName,
		Detail:     detail,
		IPAddress:  ip,
		UserAgent:  userAgent,
		Status:     status,
		ErrorMsg:   errorMsg,
	}
	return s.repo.Create(log)
}
