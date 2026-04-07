package dnsprovider

import "fmt"

// DNSProvider DNS提供商接口
type DNSProvider interface {
	// CreateDNSRecord 创建或更新DNS记录
	// domain: 完整域名 (如 api.example.com)
	// value: 记录值 (如 IP地址)
	// recordType: 记录类型 (A, AAAA, CNAME等)
	CreateDNSRecord(domain, value, recordType string) error

	// DeleteDNSRecord 删除DNS记录
	DeleteDNSRecord(domain, recordType string) error
}

// ProviderConfig DNS提供商配置
type ProviderConfig struct {
	ProviderType string // aliyun, tencent, cloudflare
	AccessKeyID     string
	AccessKeySecret string
}

// NewProvider 创建DNS提供商
func NewProvider(config ProviderConfig) (DNSProvider, error) {
	switch config.ProviderType {
	case "aliyun":
		return NewAliyunProvider(config.AccessKeyID, config.AccessKeySecret)
	case "tencent":
		return NewTencentProvider(config.AccessKeyID, config.AccessKeySecret)
	case "cloudflare":
		return NewCloudflareProvider(config.AccessKeyID)
	default:
		return nil, fmt.Errorf("不支持的DNS提供商类型: %s", config.ProviderType)
	}
}
