package dnsprovider

import (
	"fmt"
	"strings"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	alidns "github.com/alibabacloud-go/alidns-20150109/v4/client"
	"github.com/alibabacloud-go/tea/tea"
)

// AliyunProvider 阿里云DNS提供商
type AliyunProvider struct {
	client *alidns.Client
}

// NewAliyunProvider 创建阿里云DNS提供商
func NewAliyunProvider(accessKeyID, accessKeySecret string) (*AliyunProvider, error) {
	config := &openapi.Config{
		AccessKeyId:     tea.String(accessKeyID),
		AccessKeySecret: tea.String(accessKeySecret),
	}
	// 设置 endpoint
	endpoint := tea.String("alidns.aliyuncs.com")
	config.Endpoint = endpoint

	client, err := alidns.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("创建阿里云DNS客户端失败: %w", err)
	}

	return &AliyunProvider{client: client}, nil
}

// CreateDNSRecord 创建DNS记录
func (p *AliyunProvider) CreateDNSRecord(domain, value, recordType string) error {
	// 解析域名，获取主域名和RR
	rootDomain, rr := parseDomain(domain)

	// 先查询是否已存在相同记录
	existingRecordID, err := p.findRecord(rootDomain, rr, recordType)
	if err != nil {
		return fmt.Errorf("查询现有记录失败: %w", err)
	}

	if existingRecordID != "" {
		// 更新现有记录
		return p.updateRecord(existingRecordID, rr, rootDomain, value, recordType)
	}

	// 创建新记录
	return p.addRecord(rr, rootDomain, value, recordType)
}

// findRecord 查找现有记录
func (p *AliyunProvider) findRecord(domainName, rr, recordType string) (string, error) {
	req := &alidns.DescribeDomainRecordsRequest{
		DomainName:  tea.String(domainName),
		RRKeyWord:   tea.String(rr),
		TypeKeyWord: tea.String(recordType),
	}

	resp, err := p.client.DescribeDomainRecords(req)
	if err != nil {
		return "", err
	}

	if resp.Body != nil && resp.Body.DomainRecords != nil {
		for _, record := range resp.Body.DomainRecords.Record {
			if tea.StringValue(record.RR) == rr && tea.StringValue(record.Type) == recordType {
				return tea.StringValue(record.RecordId), nil
			}
		}
	}

	return "", nil
}

// addRecord 添加DNS记录
func (p *AliyunProvider) addRecord(rr, domainName, value, recordType string) error {
	req := &alidns.AddDomainRecordRequest{
		RR:         tea.String(rr),
		DomainName: tea.String(domainName),
		Type:       tea.String(recordType),
		Value:      tea.String(value),
		TTL:        tea.Int64(600),
	}

	_, err := p.client.AddDomainRecord(req)
	if err != nil {
		return fmt.Errorf("添加DNS记录失败: %w", err)
	}

	return nil
}

// updateRecord 更新DNS记录
func (p *AliyunProvider) updateRecord(recordId, rr, domainName, value, recordType string) error {
	req := &alidns.UpdateDomainRecordRequest{
		RecordId: tea.String(recordId),
		RR:       tea.String(rr),
		Type:     tea.String(recordType),
		Value:    tea.String(value),
		TTL:      tea.Int64(600),
	}

	_, err := p.client.UpdateDomainRecord(req)
	if err != nil {
		return fmt.Errorf("更新DNS记录失败: %w", err)
	}

	return nil
}

// DeleteDNSRecord 删除DNS记录
func (p *AliyunProvider) DeleteDNSRecord(domain, recordType string) error {
	rootDomain, rr := parseDomain(domain)

	// 查找记录
	recordID, err := p.findRecord(rootDomain, rr, recordType)
	if err != nil {
		return fmt.Errorf("查询记录失败: %w", err)
	}

	if recordID == "" {
		return fmt.Errorf("记录不存在")
	}

	req := &alidns.DeleteDomainRecordRequest{
		RecordId: tea.String(recordID),
	}

	_, err = p.client.DeleteDomainRecord(req)
	if err != nil {
		return fmt.Errorf("删除DNS记录失败: %w", err)
	}

	return nil
}

// parseDomain 解析域名，返回主域名和RR
// 例如: api.example.com -> (example.com, api)
// 例如: example.com -> (example.com, @)
// 例如: *.example.com -> (example.com, *)
func parseDomain(fullDomain string) (rootDomain, rr string) {
	parts := strings.Split(strings.TrimSuffix(fullDomain, "."), ".")

	// 处理常见的顶级域名
	tldCount := 2
	if len(parts) >= 2 {
		// 检查是否为常见二级顶级域名
		commonTLDs := []string{"com.cn", "net.cn", "org.cn", "co.jp", "com.hk", "com.tw"}
		domainSuffix := parts[len(parts)-2] + "." + parts[len(parts)-1]
		for _, tld := range commonTLDs {
			if domainSuffix == tld {
				tldCount = 3
				break
			}
		}
	}

	if len(parts) <= tldCount {
		// 只有主域名
		rootDomain = fullDomain
		rr = "@"
	} else {
		// 有子域名
		rootDomain = strings.Join(parts[len(parts)-tldCount:], ".")
		rr = strings.Join(parts[:len(parts)-tldCount], ".")
	}

	return rootDomain, rr
}
