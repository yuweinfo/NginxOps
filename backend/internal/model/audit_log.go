package model

import "time"

// AuditLog 审计日志实体
type AuditLog struct {
	ID         uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID     uint      `json:"userId"`
	Username   string    `json:"username" gorm:"size:50"`
	Action     string    `json:"action" gorm:"size:50"`     // LOGIN/CREATE/UPDATE/DELETE等
	Module     string    `json:"module" gorm:"size:50"`     // AUTH/SITE/CERTIFICATE等
	TargetType string    `json:"targetType" gorm:"size:50"` // 目标类型
	TargetID   uint      `json:"targetId"`                  // 目标ID
	TargetName string    `json:"targetName" gorm:"size:100"`
	Detail     string    `json:"detail" gorm:"type:text"` // JSON
	IPAddress  string    `json:"ipAddress" gorm:"size:50"`
	UserAgent  string    `json:"userAgent" gorm:"size:255"`
	Status     string    `json:"status" gorm:"size:20"` // SUCCESS/FAILURE
	ErrorMsg   string    `json:"errorMsg" gorm:"type:text"`
	CreatedAt  time.Time `json:"createdAt" gorm:"autoCreateTime"`
}

func (AuditLog) TableName() string {
	return "audit_log"
}
