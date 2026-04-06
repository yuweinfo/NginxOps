package model

import "time"

// AccessLog 访问日志实体
type AccessLog struct {
	ID          uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	RemoteAddr  string    `json:"remoteAddr" gorm:"size:50"`
	RemoteUser  string    `json:"remoteUser" gorm:"size:50"`
	TimeLocal   time.Time `json:"timeLocal"`
	Request     string    `json:"request" gorm:"size:500"`
	Method      string    `json:"method" gorm:"size:10"`
	Path        string    `json:"path" gorm:"size:500"`
	Protocol    string    `json:"protocol" gorm:"size:20"`
	Status      int       `json:"status"`
	BodyBytes   int64     `json:"bodyBytes"`
	Referer     string    `json:"referer" gorm:"size:500"`
	UserAgent   string    `json:"userAgent" gorm:"size:500"`
	RT          float64   `json:"rt"` // 响应时间(秒)
	CreatedAt   time.Time `json:"createdAt" gorm:"autoCreateTime"`
}

func (AccessLog) TableName() string {
	return "access_log"
}
