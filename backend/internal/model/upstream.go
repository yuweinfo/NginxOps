package model

import (
	"time"
)

// Upstream 负载均衡实体
type Upstream struct {
	ID            uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Name          string    `json:"name" gorm:"size:100;not null"`
	LBMode        string    `json:"lbMode" gorm:"size:20;default:round_robin"` // round_robin/ip_hash/least_conn
	HealthCheck   bool      `json:"healthCheck" gorm:"default:false"`
	CheckInterval int       `json:"checkInterval" gorm:"default:5"`
	CheckPath     string    `json:"checkPath" gorm:"size:255"`
	CheckTimeout  int       `json:"checkTimeout" gorm:"default:3"`
	Servers       string    `json:"servers" gorm:"type:text"` // JSON
	CreatedAt     time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

func (Upstream) TableName() string {
	return "upstreams"
}

// UpstreamServer 后端服务器配置
type UpstreamServer struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Weight       int    `json:"weight,omitempty"`
	MaxFails     int    `json:"maxFails,omitempty"`
	FailTimeout  int    `json:"failTimeout,omitempty"`
	Status       string `json:"status,omitempty"` // down/backup
	Backup       bool   `json:"backup,omitempty"`
}
