package model

import "time"

// DnsProvider DNS服务商配置实体
type DnsProvider struct {
	ID              uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Name            string    `json:"name" gorm:"size:100;not null"`
	ProviderType    string    `json:"providerType" gorm:"size:20;not null"` // aliyun/tencent/cloudflare
	AccessKeyID     string    `json:"accessKeyId" gorm:"size:100"`
	AccessKeySecret string    `json:"accessKeySecret" gorm:"size:255"`
	IsDefault       bool      `json:"isDefault" gorm:"default:false"`
	CreatedAt       time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt       time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

func (DnsProvider) TableName() string {
	return "dns_providers"
}
