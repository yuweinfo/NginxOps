package dnsprovider

import (
	"fmt"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	dnspod "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod/v20210323"
)

// TencentProvider 腾讯云DNS提供商
type TencentProvider struct {
	client *dnspod.Client
}

// NewTencentProvider 创建腾讯云DNS提供商
func NewTencentProvider(secretID, secretKey string) (*TencentProvider, error) {
	credential := common.NewCredential(secretID, secretKey)
	cpf := profile.NewClientProfile()
	cpf.Language = "zh-CN"

	// 设置 endpoint
	cpf.HttpProfile.Endpoint = "dnspod.tencentcloudapi.com"

	client, err := dnspod.NewClient(credential, "", cpf)
	if err != nil {
		return nil, fmt.Errorf("创建腾讯云DNS客户端失败: %w", err)
	}

	return &TencentProvider{client: client}, nil
}

// CreateDNSRecord 创建DNS记录
func (p *TencentProvider) CreateDNSRecord(domain, value, recordType string) error {
	// 解析域名，获取主域名和子域名
	rootDomain, subDomain := parseDomain(domain)

	// 获取域名ID
	domainID, err := p.getDomainID(rootDomain)
	if err != nil {
		return fmt.Errorf("获取域名ID失败: %w", err)
	}

	// 先查询是否已存在相同记录
	existingRecordID, err := p.findRecord(domainID, subDomain, recordType)
	if err != nil {
		return fmt.Errorf("查询现有记录失败: %w", err)
	}

	if existingRecordID != 0 {
		// 更新现有记录
		return p.updateRecord(existingRecordID, subDomain, rootDomain, value, recordType)
	}

	// 创建新记录
	return p.addRecord(domainID, subDomain, rootDomain, value, recordType)
}

// getDomainID 获取域名ID
func (p *TencentProvider) getDomainID(domain string) (uint64, error) {
	req := dnspod.NewDescribeDomainListRequest()
	req.Keyword = common.StringPtr(domain)

	resp, err := p.client.DescribeDomainList(req)
	if err != nil {
		return 0, err
	}

	if resp.Response != nil && resp.Response.DomainList != nil {
		for _, d := range resp.Response.DomainList {
			if tea.StringValue(d.Name) == domain {
				return tea.Uint64Value(d.DomainId), nil
			}
		}
	}

	return 0, fmt.Errorf("域名 %s 不存在或未添加到DNSPod", domain)
}

// findRecord 查找现有记录
func (p *TencentProvider) findRecord(domainID uint64, subDomain, recordType string) (uint64, error) {
	req := dnspod.NewDescribeRecordListRequest()
	req.DomainId = common.Uint64Ptr(domainID)
	req.Subdomain = common.StringPtr(subDomain)
	req.RecordType = common.StringPtr(recordType)

	resp, err := p.client.DescribeRecordList(req)
	if err != nil {
		return 0, err
	}

	if resp.Response != nil && resp.Response.RecordList != nil {
		for _, record := range resp.Response.RecordList {
			if tea.StringValue(record.Name) == subDomain && tea.StringValue(record.Type) == recordType {
				return tea.Uint64Value(record.RecordId), nil
			}
		}
	}

	return 0, nil
}

// addRecord 添加DNS记录
func (p *TencentProvider) addRecord(domainID uint64, subDomain, domain, value, recordType string) error {
	req := dnspod.NewCreateRecordRequest()
	req.DomainId = common.Uint64Ptr(domainID)
	req.SubDomain = common.StringPtr(subDomain)
	req.RecordType = common.StringPtr(recordType)
	req.Value = common.StringPtr(value)
	req.RecordLine = common.StringPtr("默认")
	req.TTL = common.Uint64Ptr(600)

	_, err := p.client.CreateRecord(req)
	if err != nil {
		return fmt.Errorf("添加DNS记录失败: %w", err)
	}

	return nil
}

// updateRecord 更新DNS记录
func (p *TencentProvider) updateRecord(recordID uint64, subDomain, domain, value, recordType string) error {
	req := dnspod.NewModifyRecordRequest()
	req.RecordId = common.Uint64Ptr(recordID)
	req.SubDomain = common.StringPtr(subDomain)
	req.RecordType = common.StringPtr(recordType)
	req.Value = common.StringPtr(value)
	req.RecordLine = common.StringPtr("默认")
	req.TTL = common.Uint64Ptr(600)

	_, err := p.client.ModifyRecord(req)
	if err != nil {
		return fmt.Errorf("更新DNS记录失败: %w", err)
	}

	return nil
}

// DeleteDNSRecord 删除DNS记录
func (p *TencentProvider) DeleteDNSRecord(domain, recordType string) error {
	rootDomain, subDomain := parseDomain(domain)

	// 获取域名ID
	domainID, err := p.getDomainID(rootDomain)
	if err != nil {
		return fmt.Errorf("获取域名ID失败: %w", err)
	}

	// 查找记录
	recordID, err := p.findRecord(domainID, subDomain, recordType)
	if err != nil {
		return fmt.Errorf("查询记录失败: %w", err)
	}

	if recordID == 0 {
		return fmt.Errorf("记录不存在")
	}

	req := dnspod.NewDeleteRecordRequest()
	req.DomainId = common.Uint64Ptr(domainID)
	req.RecordId = common.Uint64Ptr(recordID)

	_, err = p.client.DeleteRecord(req)
	if err != nil {
		return fmt.Errorf("删除DNS记录失败: %w", err)
	}

	return nil
}
