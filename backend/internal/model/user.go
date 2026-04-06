package model

import (
	"time"
)

// User 用户实体，与Java版本字段保持一致
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Username  string    `json:"username" gorm:"size:50;uniqueIndex;not null"`
	Password  string    `json:"-" gorm:"size:255;not null"`
	Email     string    `json:"email" gorm:"size:100"`
	Role      string    `json:"role" gorm:"size:20;default:USER"`
	Enabled   bool      `json:"enabled" gorm:"default:true"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

func (User) TableName() string {
	return "users"
}
