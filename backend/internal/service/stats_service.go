package service

import (
	"fmt"
	"nginxops/internal/model"
	"nginxops/internal/repository"
	"time"
)

const dashboardRankLimit = 20
const dashboardIPLocationLimit = 100
const maxDashboardRangeDays = 90

type StatsService struct {
	accessLogRepo *repository.AccessLogRepository
}

func NewStatsService() *StatsService {
	return &StatsService{
		accessLogRepo: repository.NewAccessLogRepository(),
	}
}

// GetDashboard 获取仪表盘数据（优先从内存读取，超出范围则查数据库）
func (s *StatsService) GetDashboard(start, end time.Time) map[string]interface{} {
	collector := GetLogCollector()

	// 判断时间范围是否在内存窗口内（今天）
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	useInMemory := !start.Before(today) && !end.After(now)

	if useInMemory {
		return s.getDashboardFromMemory(collector)
	}
	return s.getDashboardFromDB(start, end)
}

// getDashboardFromMemory 从内存获取实时统计（仅限当天数据）
func (s *StatsService) getDashboardFromMemory(collector *LogCollectorService) map[string]interface{} {
	statusRank := collector.GetStatusRank(dashboardRankLimit)
	ipLocations := collector.GetIPLocations(dashboardIPLocationLimit)
	regionRank := collector.GetRegionRank(dashboardRankLimit)
	ipTopRank := collector.GetIPTopRank(dashboardRankLimit)
	hostRank := collector.GetHostRank(dashboardRankLimit)
	refererRank := collector.GetRefererRank(dashboardRankLimit)
	pathRank := collector.GetPathRank(dashboardRankLimit)
	resourceTypeRank := collector.GetResourceTypeRank(dashboardRankLimit)
	browserRank := collector.GetBrowserRank(dashboardRankLimit)
	deviceTypeRank := collector.GetDeviceTypeRank(dashboardRankLimit)
	osRank := collector.GetOSRank(dashboardRankLimit)
	userAgentRank := collector.GetUserAgentRank(dashboardRankLimit)

	return map[string]interface{}{
		"statusRank":       statusRank,
		"ipLocations":      ipLocations,
		"ipRegionRank":     regionRank,
		"ipTopRank":        ipTopRank,
		"hostRank":         hostRank,
		"refererRank":      refererRank,
		"pathRank":         pathRank,
		"resourceTypeRank": resourceTypeRank,
		"browserRank":      browserRank,
		"deviceTypeRank":    deviceTypeRank,
		"osRank":           osRank,
		"userAgentRank":    userAgentRank,
	}
}

// getDashboardFromDB 从数据库获取统计（支持时间范围查询）
func (s *StatsService) getDashboardFromDB(start, end time.Time) map[string]interface{} {
	result := make(map[string]interface{})

	// 状态码排行
	if statusRank, err := s.accessLogRepo.GetStatusRank(start, end, dashboardRankLimit); err == nil {
		result["statusRank"] = rankItemsToResponse(statusRank)
	} else {
		result["statusRank"] = []interface{}{}
	}

	// IP 地理位置排行
	if ipLocations, err := s.accessLogRepo.CountByLocation(start, end, dashboardIPLocationLimit); err == nil {
		result["ipLocations"] = ipLocationCountsToResponse(ipLocations)
	} else {
		result["ipLocations"] = []interface{}{}
	}

	// 地区排名
	if regionRank, err := s.accessLogRepo.CountByRegion(start, end, dashboardRankLimit); err == nil {
		result["ipRegionRank"] = regionCountsToResponse(regionRank)
	} else {
		result["ipRegionRank"] = []interface{}{}
	}

	// IP Top 排行
	if ipTopRank, err := s.accessLogRepo.GetIPTopRank(start, end, dashboardRankLimit); err == nil {
		result["ipTopRank"] = ipLocationCountsToIPTopResponse(ipTopRank)
	} else {
		result["ipTopRank"] = []interface{}{}
	}

	// Host 排行
	if hostRank, err := s.accessLogRepo.GetHostRank(start, end, dashboardRankLimit); err == nil {
		result["hostRank"] = rankItemsToResponse(hostRank)
	} else {
		result["hostRank"] = []interface{}{}
	}

	// Referer 排行
	if refererRank, err := s.accessLogRepo.GetRefererRank(start, end, dashboardRankLimit); err == nil {
		result["refererRank"] = rankItemsToResponse(refererRank)
	} else {
		result["refererRank"] = []interface{}{}
	}

	// URL Path 排行
	if pathRank, err := s.accessLogRepo.GetPathRank(start, end, dashboardRankLimit); err == nil {
		result["pathRank"] = rankItemsToResponse(pathRank)
	} else {
		result["pathRank"] = []interface{}{}
	}

	// 资源类型排行
	if resourceTypeRank, err := s.accessLogRepo.GetResourceTypeRank(start, end, dashboardRankLimit); err == nil {
		result["resourceTypeRank"] = rankItemsToResponse(resourceTypeRank)
	} else {
		result["resourceTypeRank"] = []interface{}{}
	}

	// 浏览器、设备类型、操作系统、User-Agent 排行
	// 这些需要从 user_agent 解析，数据库中没有独立字段
	// 使用内存数据（当天的），或从 user_agent 聚合
	s.fillUserAgentRanksFromDB(start, end, result)

	return result
}

// fillUserAgentRanksFromDB 从数据库 user_agent 字段聚合客户端分析数据
func (s *StatsService) fillUserAgentRanksFromDB(start, end time.Time, result map[string]interface{}) {
	// 获取 user_agent 排行
	uaRanks, err := s.accessLogRepo.GetUserAgentRank(start, end, dashboardRankLimit*5)
	if err != nil {
		result["browserRank"] = []interface{}{}
		result["deviceTypeRank"] = []interface{}{}
		result["osRank"] = []interface{}{}
		result["userAgentRank"] = rankItemsToResponse(nil)
		return
	}

	// 在内存中解析 user_agent 并分类
	browserCounts := make(map[string]int64)
	osCounts := make(map[string]int64)
	deviceTypeCounts := make(map[string]int64)

	collector := GetLogCollector()
	for _, item := range uaRanks {
		ua := item.Name
		if ua == "" {
			continue
		}
		count := item.Count
		browser := collector.ParseBrowser(ua)
		os := collector.ParseOS(ua)
		deviceType := collector.ParseDeviceType(ua)

		browserCounts[browser] += count
		osCounts[os] += count
		deviceTypeCounts[deviceType] += count
	}

	result["browserRank"] = countsToRankResponse(browserCounts, dashboardRankLimit)
	result["osRank"] = countsToRankResponse(osCounts, dashboardRankLimit)
	result["deviceTypeRank"] = countsToRankResponse(deviceTypeCounts, dashboardRankLimit)
	result["userAgentRank"] = rankItemsToResponse(uaRanks[:min(len(uaRanks), dashboardRankLimit)])
}

// rankItemsToResponse 将 RankItem 列表转为前端需要的格式
func rankItemsToResponse(items []repository.RankItem) []map[string]interface{} {
	var total int64
	for _, item := range items {
		total += item.Count
	}

	result := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		percent := float64(0)
		if total > 0 {
			percent = float64(item.Count) / float64(total) * 100
		}
		result = append(result, map[string]interface{}{
			"name":    item.Name,
			"count":   item.Count,
			"percent": percent,
		})
	}
	return result
}

// ipLocationCountsToResponse 将 IpLocationCount 转为地图数据格式
func ipLocationCountsToResponse(items []repository.IpLocationCount) []map[string]interface{} {
	// 按国家聚合
	type countryStat struct {
		Country  string
		Lat      float64
		Lon      float64
		Requests int64
	}
	countryMap := make(map[string]*countryStat)
	for _, item := range items {
		if c, ok := countryMap[item.Country]; ok {
			c.Requests += item.Requests
		} else {
			countryMap[item.Country] = &countryStat{
				Country:  item.Country,
				Lat:      item.Lat,
				Lon:      item.Lon,
				Requests: item.Requests,
			}
		}
	}

	result := make([]map[string]interface{}, 0)
	for _, c := range countryMap {
		if c.Lat != 0 && c.Lon != 0 {
			result = append(result, map[string]interface{}{
				"name":    c.Country,
				"value":   []float64{c.Lon, c.Lat, float64(c.Requests)},
				"country": c.Country,
			})
		}
	}
	return result
}

// regionCountsToResponse 将 RegionCount 转为前端格式
func regionCountsToResponse(items []repository.RegionCount) []map[string]interface{} {
	var total int64
	for _, item := range items {
		total += item.Count
	}

	result := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		percent := float64(0)
		if total > 0 {
			percent = float64(item.Count) / float64(total) * 100
		}
		result = append(result, map[string]interface{}{
			"country": item.City,
			"count":   item.Count,
			"percent": percent,
		})
	}
	return result
}

// ipLocationCountsToIPTopResponse 将 IpLocationCount 转为 IP Top 排行格式
func ipLocationCountsToIPTopResponse(items []repository.IpLocationCount) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		result = append(result, map[string]interface{}{
			"ip":       item.IP,
			"region":   item.Country,
			"requests": item.Requests,
		})
	}
	return result
}

// countsToRankResponse 将 map[string]int64 转为排名格式
func countsToRankResponse(counts map[string]int64, limit int) []map[string]interface{} {
	type entry struct {
		name  string
		count int64
	}
	entries := make([]entry, 0, len(counts))
	for name, count := range counts {
		entries = append(entries, entry{name: name, count: count})
	}

	// 排序
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[j].count > entries[i].count {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}

	if limit > len(entries) {
		limit = len(entries)
	}

	var total int64
	for _, e := range entries {
		total += e.count
	}

	result := make([]map[string]interface{}, 0, limit)
	for i := 0; i < limit; i++ {
		percent := float64(0)
		if total > 0 {
			percent = float64(entries[i].count) / float64(total) * 100
		}
		result = append(result, map[string]interface{}{
			"name":    entries[i].name,
			"count":   entries[i].count,
			"percent": percent,
		})
	}
	return result
}

// ValidateDashboardRange 校验时间范围，返回校验后的 start, end 和错误信息
func ValidateDashboardRange(start, end time.Time) (time.Time, time.Time, error) {
	if start.IsZero() {
		end = time.Now()
		start = end.Add(-24 * time.Hour)
	}

	if end.IsZero() {
		end = time.Now()
	}

	if end.Before(start) {
		return time.Time{}, time.Time{}, fmt.Errorf("结束时间不能早于开始时间")
	}

	maxRange := time.Duration(maxDashboardRangeDays) * 24 * time.Hour
	if end.Sub(start) > maxRange {
		return time.Time{}, time.Time{}, fmt.Errorf("查询时间范围不能超过 %d 天", maxDashboardRangeDays)
	}

	return start, end, nil
}

// QueryLogs 查询访问日志（从数据库分页读取）
func (s *StatsService) QueryLogs(start, end time.Time, ip string, page, size int) ([]model.AccessLog, int64, error) {
	return s.accessLogRepo.FindPage(page, size, start, end, ip)
}
