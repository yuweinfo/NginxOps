package model

import (
	"time"
)

// Certificate 证书实体
type Certificate struct {
	ID        uint       `json:"id" gorm:"primaryKey;autoIncrement"`
	Domain    string     `json:"domain" gorm:"size:255;not null"`
	Issuer    string     `json:"issuer" gorm:"size:100"`
	IssuedAt  *time.Time `json:"issuedAt"`
	ExpiresAt *time.Time `json:"expiresAt"`
	Status    string     `json:"status" gorm:"size:20;default:pending"` // pending/valid/expired/failed
	AutoRenew bool       `json:"autoRenew" gorm:"default:false"`
	CertPath  string     `json:"certPath" gorm:"size:255"`
	KeyPath   string     `json:"keyPath" gorm:"size:255"`
	CreatedAt time.Time  `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time  `json:"updatedAt" gorm:"autoUpdateTime"`
}

func (Certificate) TableName() string {
	return "certificates"
}
