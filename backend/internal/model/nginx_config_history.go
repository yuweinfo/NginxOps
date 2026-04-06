package model

import (
	"time"
)

// NginxConfigHistory 配置历史实体
type NginxConfigHistory struct {
	ID         uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	ConfigName string    `json:"configName" gorm:"size:100"`
	ConfigType string    `json:"configType" gorm:"size:20"` // nginx/confd
	FilePath   string    `json:"filePath" gorm:"size:255"`
	Content    string    `json:"content" gorm:"type:text"`
	Remark     string    `json:"remark" gorm:"size:255"`
	CreatedAt  time.Time `json:"createdAt" gorm:"autoCreateTime"`
}

func (NginxConfigHistory) TableName() string {
	return "nginx_config_history"
}
