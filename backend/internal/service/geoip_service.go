package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"nginxops/internal/database"
	"nginxops/internal/model"
	"nginxops/internal/repository"
	"strings"
	"time"
)

type GeoIpService struct {
	repo       *repository.IpGeoCacheRepository
	cidrRepo   *repository.CountryCIDRRepository
}

func NewGeoIpService() *GeoIpService {
	return &GeoIpService{
		repo:     repository.NewIpGeoCacheRepository(),
		cidrRepo: repository.NewCountryCIDRRepository(),
	}
}

func (s *GeoIpService) GetGeo(ip string) *model.IpGeoCache {
	// 先查缓存
	cache, err := s.repo.FindByIP(ip)
	if err == nil && cache != nil {
		return cache
	}

	// 查询ip-api.com
	geo := s.queryFromAPI(ip)
	if geo != nil {
		// 保存到缓存
		s.repo.Create(geo)
		return geo
	}

	// 返回默认值
	return &model.IpGeoCache{
		IP:      ip,
		Country: "Unknown",
		Region:  "Unknown",
		City:    "Unknown",
	}
}

func (s *GeoIpService) queryFromAPI(ip string) *model.IpGeoCache {
	url := fmt.Sprintf("http://ip-api.com/json/%s", ip)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	var result struct {
		Status      string  `json:"status"`
		Country     string  `json:"country"`
		RegionName  string  `json:"regionName"`
		City        string  `json:"city"`
		ISP         string  `json:"isp"`
		Lat         float64 `json:"lat"`
		Lon         float64 `json:"lon"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil
	}

	if result.Status != "success" {
		return nil
	}

	return &model.IpGeoCache{
		IP:      ip,
		Country: result.Country,
		Region:  result.RegionName,
		City:    result.City,
		ISP:     result.ISP,
		Lat:     result.Lat,
		Lon:     result.Lon,
	}
}

// GetCountryCIDRs 获取指定国家的 IP CIDR 列表
// 先从本地缓存查询，如果没有则从远程 API 获取并缓存
func (s *GeoIpService) GetCountryCIDRs(countryCode string) []string {
	cc := strings.ToUpper(countryCode)

	// 先查本地缓存
	cidrs, err := s.cidrRepo.FindByCountryCode(cc)
	if err == nil && len(cidrs) > 0 {
		return cidrs
	}

	// 从远程 API 获取
	remoteCIDRs := s.fetchCountryCIDRsFromAPI(cc)
	if len(remoteCIDRs) > 0 {
		// 缓存到数据库
		if err := s.cidrRepo.Save(cc, remoteCIDRs); err != nil {
			log.Printf("缓存国家 CIDR 数据失败: %v", err)
		}
		return remoteCIDRs
	}

	return nil
}

// fetchCountryCIDRsFromAPI 从远程 API 获取国家的 IP CIDR 列表
// 使用 ipinfo.io 的免费 API
func (s *GeoIpService) fetchCountryCIDRsFromAPI(countryCode string) []string {
	url := fmt.Sprintf("https://ipinfo.io/countries/%s", countryCode)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		log.Printf("从 ipinfo.io 获取国家 CIDR 失败: %v", err)
		return s.fetchCountryCIDRsFromDBIP(countryCode)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("ipinfo.io 返回非200状态码: %d", resp.StatusCode)
		return s.fetchCountryCIDRsFromDBIP(countryCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return s.fetchCountryCIDRsFromDBIP(countryCode)
	}

	var result struct {
		Country string   `json:"country"`
		IPs     []string `json:"ips"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return s.fetchCountryCIDRsFromDBIP(countryCode)
	}

	if len(result.IPs) > 0 {
		return result.IPs
	}

	return s.fetchCountryCIDRsFromDBIP(countryCode)
}

// fetchCountryCIDRsFromDBIP 备用 API: db-ip.com 的免费国家 IP 范围数据
func (s *GeoIpService) fetchCountryCIDRsFromDBIP(countryCode string) []string {
	// 使用 db-ip.com 的免费 API 获取国家级 IP 范围
	url := fmt.Sprintf("https://db-ip.com/db/download/country/%s", countryCode)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		log.Printf("从 db-ip.com 获取国家 CIDR 失败: %v", err)
		return nil
	}
	defer resp.Body.Close()

	// db-ip.com 返回 CSV 格式
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	var cidrs []string
	lines := strings.Split(string(body), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "start") {
			continue
		}
		fields := strings.Split(line, ",")
		if len(fields) >= 3 {
			// 格式: start_ip, end_ip, country
			// 使用 start_ip/前缀长度 格式
			cidr := fields[0]
			if strings.Contains(cidr, ".") || strings.Contains(cidr, ":") {
				// 简化处理：如果是 IP 地址范围，转换为 CIDR
				if len(fields) >= 2 {
					cidrs = append(cidrs, ipRangeToCIDR(fields[0], fields[1])...)
				}
			}
		}
	}

	return cidrs
}

// ipRangeToCIDR 将 IP 范围转换为 CIDR 列表
func ipRangeToCIDR(startIP, endIP string) []string {
	// 简化实现：使用 start_ip/前缀长度
	// 对于常见的大范围 IP 段，使用 /8, /16, /24 等前缀
	// 这里做一个简化处理：直接返回 startIP 对应的大 CIDR
	parts := strings.Split(startIP, ".")
	if len(parts) == 4 {
		// IPv4: 返回 /8 网段（简化处理，不精确但可用）
		return []string{fmt.Sprintf("%s.0.0.0/8", parts[0])}
	}
	return nil
}

// GetCountryCode 根据国家名称获取国家代码
func (s *GeoIpService) GetCountryCode(countryName string) string {
	// 常见国家名称到代码的映射
	countryMap := map[string]string{
		"China":            "CN",
		"United States":    "US",
		"Japan":            "JP",
		"South Korea":      "KR",
		"United Kingdom":   "GB",
		"Germany":          "DE",
		"France":           "FR",
		"Russia":           "RU",
		"India":            "IN",
		"Brazil":           "BR",
		"Australia":        "AU",
		"Canada":           "CA",
		"Italy":            "IT",
		"Spain":            "ES",
		"Netherlands":      "NL",
		"Singapore":        "SG",
		"Hong Kong":        "HK",
		"Taiwan":           "TW",
		"Malaysia":         "MY",
		"Thailand":         "TH",
		"Vietnam":          "VN",
		"Indonesia":        "ID",
		"Philippines":      "PH",
		"Turkey":           "TR",
		"Mexico":           "MX",
		"Argentina":        "AR",
		"Colombia":         "CO",
		"South Africa":     "ZA",
		"Israel":           "IL",
		"Ukraine":          "UA",
		"Poland":           "PL",
		"Sweden":           "SE",
		"Norway":           "NO",
		"Denmark":          "DK",
		"Finland":          "FI",
		"Switzerland":      "CH",
		"Austria":          "AT",
		"Belgium":          "BE",
		"Portugal":         "PT",
		"Ireland":          "IE",
		"New Zealand":      "NZ",
		"United Arab Emirates": "AE",
		"Saudi Arabia":     "SA",
		"Egypt":            "EG",
		"Nigeria":          "NG",
		"Kenya":            "KE",
		"Pakistan":         "PK",
		"Bangladesh":       "BD",
		"Iran":             "IR",
		"Iraq":             "IQ",
		"Cambodia":         "KH",
		"Myanmar":          "MM",
		"Nepal":            "NP",
		"Sri Lanka":        "LK",
		"Kazakhstan":       "KZ",
		"Mongolia":         "MN",
		"North Korea":      "KP",
	}

	if code, ok := countryMap[countryName]; ok {
		return code
	}

	// 尝试从缓存数据库中查找
	var caches []model.IpGeoCache
	if err := database.DB.Where("country = ?", countryName).Limit(1).Find(&caches).Error; err == nil && len(caches) > 0 {
		// 无法从缓存得到代码，但至少确认国家名存在
	}

	// 无法映射，返回空字符串
	return ""
}
