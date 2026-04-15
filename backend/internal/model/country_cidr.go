package model

import "time"

// CountryCIDR 国家IP段缓存
type CountryCIDR struct {
	ID          uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	CountryCode string    `json:"countryCode" gorm:"size:10;not null;uniqueIndex"`
	CIDRs       string    `json:"cidrs" gorm:"type:text"` // JSON 数组存储 CIDR 列表
	UpdatedAt   time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

func (CountryCIDR) TableName() string {
	return "country_cidrs"
}
