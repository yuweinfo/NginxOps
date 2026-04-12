package model

import (
	"time"
)

// IPBlacklist 全局 IP 黑名单
type IPBlacklist struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	IPAddress string    `json:"ipAddress" gorm:"size:100;not null"`
	Note      string    `json:"note" gorm:"size:255"`
	Enabled   bool      `json:"enabled" gorm:"default:true"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

func (IPBlacklist) TableName() string {
	return "ip_blacklist"
}

// GeoRule 全局 Geo 封锁规则
type GeoRule struct {
	ID          uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	CountryCode string    `json:"countryCode" gorm:"size:10;not null"`
	Action      string    `json:"action" gorm:"size:10;not null;default:block"` // allow/block
	Note        string    `json:"note" gorm:"size:255"`
	Enabled     bool      `json:"enabled" gorm:"default:true"`
	CreatedAt   time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

func (GeoRule) TableName() string {
	return "geo_rules"
}

// AccessControlSettings 全局访问控制设置
type AccessControlSettings struct {
	ID                 uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	GeoEnabled         bool      `json:"geoEnabled" gorm:"default:false"`
	IPBlacklistEnabled bool      `json:"ipBlacklistEnabled" gorm:"default:false"`
	DefaultAction      string    `json:"defaultAction" gorm:"size:10;default:allow"` // allow/block
	UpdatedAt          time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

func (AccessControlSettings) TableName() string {
	return "access_control_settings"
}

// SiteIPBlacklist 站点专属 IP 黑名单
type SiteIPBlacklist struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	SiteID    uint      `json:"siteId" gorm:"not null"`
	IPAddress string    `json:"ipAddress" gorm:"size:100;not null"`
	Note      string    `json:"note" gorm:"size:255"`
	Enabled   bool      `json:"enabled" gorm:"default:true"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

func (SiteIPBlacklist) TableName() string {
	return "site_ip_blacklist"
}

// SiteGeoRule 站点专属 Geo 规则
type SiteGeoRule struct {
	ID          uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	SiteID      uint      `json:"siteId" gorm:"not null"`
	CountryCode string    `json:"countryCode" gorm:"size:10;not null"`
	Action      string    `json:"action" gorm:"size:10;not null;default:block"`
	Note        string    `json:"note" gorm:"size:255"`
	Enabled     bool      `json:"enabled" gorm:"default:true"`
	CreatedAt   time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

func (SiteGeoRule) TableName() string {
	return "site_geo_rules"
}
