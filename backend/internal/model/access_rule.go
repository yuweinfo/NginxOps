package model

import (
	"time"
)

// AccessRule 访问控制规则 - 每条规则是独立的、可复用的单元
type AccessRule struct {
	ID          uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string    `json:"name" gorm:"size:100;not null"`
	Description string    `json:"description" gorm:"size:255"`
	Enabled     bool      `json:"enabled" gorm:"default:true"`
	CreatedAt   time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
	Items       []AccessRuleItem `json:"items" gorm:"foreignKey:RuleID"`
}

func (AccessRule) TableName() string {
	return "access_rules"
}

// AccessRuleItem 规则条目 - 一条规则可以包含多个 IP 条目和 Geo 条目
type AccessRuleItem struct {
	ID          uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	RuleID      uint      `json:"ruleId" gorm:"not null;index"`
	ItemType    string    `json:"itemType" gorm:"size:20;not null"` // ip / geo
	IPAddress   string    `json:"ipAddress" gorm:"size:100"`        // 当 ItemType=ip 时使用
	CountryCode string    `json:"countryCode" gorm:"size:10"`       // 当 ItemType=geo 时使用
	Action      string    `json:"action" gorm:"size:10;default:block"` // allow / block（IP 和 Geo 均使用）
	Note        string    `json:"note" gorm:"size:255"`
	CreatedAt   time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

func (AccessRuleItem) TableName() string {
	return "access_rule_items"
}

// SiteAccessRule 站点-规则关联表
type SiteAccessRule struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	SiteID    uint      `json:"siteId" gorm:"not null;uniqueIndex:idx_site_rule"`
	RuleID    uint      `json:"ruleId" gorm:"not null;uniqueIndex:idx_site_rule"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
}

func (SiteAccessRule) TableName() string {
	return "site_access_rules"
}
