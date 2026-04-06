package model

import "time"

// IpGeoCache IP地理位置缓存实体
type IpGeoCache struct {
	IP        string    `json:"ip" gorm:"primaryKey;size:50"`
	Country   string    `json:"country" gorm:"size:100"`
	Region    string    `json:"region" gorm:"size:100"`
	City      string    `json:"city" gorm:"size:100"`
	ISP       string    `json:"isp" gorm:"size:100"`
	Lat       float64   `json:"lat"`
	Lon       float64   `json:"lon"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

func (IpGeoCache) TableName() string {
	return "ip_geo_cache"
}
