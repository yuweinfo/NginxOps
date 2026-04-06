package service

import (
	"nginxops/internal/model"
	"nginxops/internal/repository"
	"time"
)

type StatsService struct {
	accessLogRepo *repository.AccessLogRepository
}

func NewStatsService() *StatsService {
	return &StatsService{
		accessLogRepo: repository.NewAccessLogRepository(),
	}
}

func (s *StatsService) GetDashboard() map[string]interface{} {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	pv, _ := s.accessLogRepo.CountPV(today, now)
	uv, _ := s.accessLogRepo.CountUV(today, now)
	statusCounts, _ := s.accessLogRepo.CountByStatus(today, now)

	return map[string]interface{}{
		"pv":     pv,
		"uv":     uv,
		"status": statusCounts,
		"qps":    float64(pv) / 86400.0,
	}
}

func (s *StatsService) QueryLogs(start, end time.Time, ip string, page, size int) ([]model.AccessLog, int64, error) {
	return s.accessLogRepo.FindPage(page, size, start, end, ip)
}
