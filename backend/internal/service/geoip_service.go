package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"nginxops/internal/model"
	"nginxops/internal/repository"
	"time"
)

type GeoIpService struct {
	repo *repository.IpGeoCacheRepository
}

func NewGeoIpService() *GeoIpService {
	return &GeoIpService{
		repo: repository.NewIpGeoCacheRepository(),
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
