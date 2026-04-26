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

// GetDashboard 获取仪表盘数据（优先从内存读取）
func (s *StatsService) GetDashboard(start, end time.Time) map[string]interface{} {
	collector := GetLogCollector()

	// 从内存获取实时统计
	statusRank := collector.GetStatusRank(20)
	ipLocations := collector.GetIPLocations(100)
	regionRank := collector.GetRegionRank(20)
	ipTopRank := collector.GetIPTopRank(20)
	hostRank := collector.GetHostRank(20)
	refererRank := collector.GetRefererRank(20)
	pathRank := collector.GetPathRank(20)
	resourceTypeRank := collector.GetResourceTypeRank(20)
	browserRank := collector.GetBrowserRank(20)
	deviceTypeRank := collector.GetDeviceTypeRank(20)
	osRank := collector.GetOSRank(20)
	userAgentRank := collector.GetUserAgentRank(20)

	return map[string]interface{}{
		"statusRank":         statusRank,
		"ipLocations":        ipLocations,
		"ipRegionRank":       regionRank,
		"ipTopRank":          ipTopRank,
		"hostRank":           hostRank,
		"refererRank":        refererRank,
		"pathRank":           pathRank,
		"resourceTypeRank":   resourceTypeRank,
		"browserRank":        browserRank,
		"deviceTypeRank":     deviceTypeRank,
		"osRank":             osRank,
		"userAgentRank":      userAgentRank,
	}
}

// QueryLogs 查询访问日志（从数据库分页读取）
func (s *StatsService) QueryLogs(start, end time.Time, ip string, page, size int) ([]model.AccessLog, int64, error) {
	return s.accessLogRepo.FindPage(page, size, start, end, ip)
}
