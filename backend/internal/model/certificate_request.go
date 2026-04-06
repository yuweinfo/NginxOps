package model

import "time"

// CertificateRequest 证书申请记录
type CertificateRequest struct {
	ID             uint       `json:"id" gorm:"primaryKey;autoIncrement"`
	CertificateID  uint       `json:"certificateId"`
	Status         string     `json:"status" gorm:"size:20;default:pending"` // pending/validating/issuing/completed/failed
	ChallengeType  string     `json:"challengeType" gorm:"size:20;default:dns-01"`
	DNSProviderID  uint       `json:"dnsProviderId"`
	DNSRecordName  string     `json:"dnsRecordName" gorm:"size:255"`
	DNSRecordValue string     `json:"dnsRecordValue" gorm:"size:255"`
	ErrorMessage   string     `json:"errorMessage" gorm:"type:text"`
	StartedAt      *time.Time `json:"startedAt"`
	CompletedAt    *time.Time `json:"completedAt"`
	CreatedAt      time.Time  `json:"createdAt" gorm:"autoCreateTime"`
}

func (CertificateRequest) TableName() string {
	return "certificate_requests"
}
