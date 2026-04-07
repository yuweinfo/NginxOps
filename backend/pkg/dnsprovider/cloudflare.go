package dnsprovider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// CloudflareProvider Cloudflare DNS提供商
type CloudflareProvider struct {
	apiToken string
	client   *http.Client
}

// NewCloudflareProvider 创建Cloudflare DNS提供商
func NewCloudflareProvider(apiToken string) (*CloudflareProvider, error) {
	if apiToken == "" {
		return nil, fmt.Errorf("API Token 不能为空")
	}

	return &CloudflareProvider{
		apiToken: apiToken,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// Cloudflare API 响应结构
type cfResponse struct {
	Success  bool            `json:"success"`
	Errors   []cfError       `json:"errors"`
	Result   json.RawMessage `json:"result"`
	ResultInfo *cfResultInfo `json:"result_info,omitempty"`
}

type cfError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type cfResultInfo struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	TotalPages int `json:"total_pages"`
	Count      int `json:"count"`
	TotalCount int `json:"total_count"`
}

type cfZone struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type cfDNSRecord struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	TTL     int    `json:"ttl"`
	Proxied bool   `json:"proxied"`
}

// CreateDNSRecord 创建DNS记录
func (p *CloudflareProvider) CreateDNSRecord(domain, value, recordType string) error {
	// 解析域名，获取主域名
	rootDomain, _ := parseDomain(domain)

	// 获取 Zone ID
	zoneID, err := p.getZoneID(rootDomain)
	if err != nil {
		return fmt.Errorf("获取 Zone ID 失败: %w", err)
	}

	// 查找现有记录
	existingRecord, err := p.findRecord(zoneID, domain, recordType)
	if err != nil {
		return fmt.Errorf("查询现有记录失败: %w", err)
	}

	if existingRecord != nil {
		// 更新现有记录
		return p.updateRecord(zoneID, existingRecord.ID, domain, value, recordType)
	}

	// 创建新记录
	return p.addRecord(zoneID, domain, value, recordType)
}

// getZoneID 获取Zone ID
func (p *CloudflareProvider) getZoneID(domain string) (string, error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones?name=%s", domain)

	resp, err := p.doRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	var zones []cfZone
	if err := p.parseResult(resp, &zones); err != nil {
		return "", err
	}

	if len(zones) == 0 {
		return "", fmt.Errorf("域名 %s 不存在或未添加到 Cloudflare", domain)
	}

	return zones[0].ID, nil
}

// findRecord 查找现有记录
func (p *CloudflareProvider) findRecord(zoneID, domain, recordType string) (*cfDNSRecord, error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records?type=%s&name=%s", zoneID, recordType, domain)

	resp, err := p.doRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	var records []cfDNSRecord
	if err := p.parseResult(resp, &records); err != nil {
		return nil, err
	}

	if len(records) > 0 {
		return &records[0], nil
	}

	return nil, nil
}

// addRecord 添加DNS记录
func (p *CloudflareProvider) addRecord(zoneID, domain, value, recordType string) error {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records", zoneID)

	record := map[string]interface{}{
		"type":    recordType,
		"name":    domain,
		"content": value,
		"ttl":     600,
		"proxied": false,
	}

	body, err := json.Marshal(record)
	if err != nil {
		return err
	}

	_, err = p.doRequest("POST", url, body)
	return err
}

// updateRecord 更新DNS记录
func (p *CloudflareProvider) updateRecord(zoneID, recordID, domain, value, recordType string) error {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", zoneID, recordID)

	record := map[string]interface{}{
		"type":    recordType,
		"name":    domain,
		"content": value,
		"ttl":     600,
		"proxied": false,
	}

	body, err := json.Marshal(record)
	if err != nil {
		return err
	}

	_, err = p.doRequest("PUT", url, body)
	return err
}

// DeleteDNSRecord 删除DNS记录
func (p *CloudflareProvider) DeleteDNSRecord(domain, recordType string) error {
	rootDomain, _ := parseDomain(domain)

	// 获取 Zone ID
	zoneID, err := p.getZoneID(rootDomain)
	if err != nil {
		return fmt.Errorf("获取 Zone ID 失败: %w", err)
	}

	// 查找记录
	record, err := p.findRecord(zoneID, domain, recordType)
	if err != nil {
		return fmt.Errorf("查询记录失败: %w", err)
	}

	if record == nil {
		return fmt.Errorf("记录不存在")
	}

	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", zoneID, record.ID)
	_, err = p.doRequest("DELETE", url, nil)
	return err
}

// doRequest 执行HTTP请求
func (p *CloudflareProvider) doRequest(method, url string, body []byte) (*cfResponse, error) {
	var req *http.Request
	var err error

	if body != nil {
		req, err = http.NewRequest(method, url, strings.NewReader(string(body)))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+p.apiToken)
	req.Header.Set("Content-Type", "application/json")

	httpResp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}

	var resp cfResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if !resp.Success {
		if len(resp.Errors) > 0 {
			return nil, fmt.Errorf("Cloudflare API 错误: %s", resp.Errors[0].Message)
		}
		return nil, fmt.Errorf("Cloudflare API 请求失败")
	}

	return &resp, nil
}

// parseResult 解析API结果
func (p *CloudflareProvider) parseResult(resp *cfResponse, v interface{}) error {
	if resp.Result == nil {
		return nil
	}
	return json.Unmarshal(resp.Result, v)
}
