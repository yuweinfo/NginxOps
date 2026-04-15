package model

import (
	"time"
)

// Site 站点实体，与Java版本字段保持一致
type Site struct {
	ID                 uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	FileName           string    `json:"fileName" gorm:"size:100"`
	Domain             string    `json:"domain" gorm:"size:255;not null"`
	Port               int       `json:"port" gorm:"default:80"`
	SiteType           string    `json:"siteType" gorm:"size:20;default:proxy"` // static/proxy/loadbalance
	RootDir            string    `json:"rootDir" gorm:"size:255"`
	Locations          string    `json:"locations" gorm:"type:text"`            // JSON
	UpstreamID         *uint     `json:"upstreamId"`                            // 关联已定义的 upstream
	UpstreamServers    string    `json:"upstreamServers" gorm:"type:text"`      // JSON (站点自己定义的后端服务器)
	SSLEnabled         bool      `json:"sslEnabled" gorm:"default:false"`
	CertID             *uint     `json:"certId"`
	ForceHttps         bool      `json:"forceHttps" gorm:"default:false"`
	Gzip               bool      `json:"gzip" gorm:"default:false"`
	Cache              bool      `json:"cache" gorm:"default:false"`
	MaxBodySize        int       `json:"maxBodySize" gorm:"default:200"`        // 最大上传大小（MB），默认200MB
	AccessControlMode  string    `json:"accessControlMode" gorm:"size:20;default:custom"` // custom (新设计：通过规则关联)
	Enabled            bool      `json:"enabled" gorm:"default:true"`
	Config             string    `json:"config" gorm:"type:text"`
	CreatedAt          time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt          time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

func (Site) TableName() string {
	return "sites"
}
