package service

import (
	"nginxops/internal/model"
	"nginxops/internal/repository"
	"time"
)

type StatsService struct {
	accessLogRepo *repository.AccessLogRepository
	siteRepo      *repository.SiteRepository
}

func NewStatsService() *StatsService {
	return &StatsService{
		accessLogRepo: repository.NewAccessLogRepository(),
		siteRepo:      repository.NewSiteRepository(),
	}
}

// GetDashboard 获取仪表盘数据（优先从内存读取）
func (s *StatsService) GetDashboard() map[string]interface{} {
	collector := GetLogCollector()

	// 从内存获取实时统计
	qps := collector.GetQPS()
	pvToday := collector.GetTodayPV()
	uvToday := collector.GetTodayUV()
	bandwidth := collector.GetBandwidth()
	statusDistribution := collector.GetStatusDistribution()
	hourlyTrend := collector.GetHourlyTrend()
	ipLocations := collector.GetIPLocations(100)
	regionRank := collector.GetRegionRank(10)

	// 活跃站点数（从数据库获取，因为变化频率低）
	sites, _ := s.siteRepo.FindByEnabled(true)
	activeSites := len(sites)
	collector.SetActiveSites(activeSites)

	return map[string]interface{}{
		"qps":                qps,
		"bandwidth":          bandwidth,
		"pvToday":            pvToday,
		"uvToday":            uvToday,
		"activeSites":        activeSites,
		"hourlyTrend":        hourlyTrend,
		"statusDistribution": statusDistribution,
		"ipLocations":        ipLocations,
		"ipRegionRank":       regionRank,
	}
}

// QueryLogs 查询访问日志（从数据库分页读取）
func (s *StatsService) QueryLogs(start, end time.Time, ip string, page, size int) ([]model.AccessLog, int64, error) {
	return s.accessLogRepo.FindPage(page, size, start, end, ip)
}
