package handler

import (
	"io/ioutil"
	"net"
	"net/http"
	"nginxops/internal/model"
	"nginxops/internal/service"
	"nginxops/pkg/dnsprovider"
	"nginxops/pkg/response"
	"strings"

	"github.com/gin-gonic/gin"
)

type NetworkHandler struct {
	dnsService *service.DnsProviderService
}

func NewNetworkHandler() *NetworkHandler {
	return &NetworkHandler{
		dnsService: service.NewDnsProviderService(),
	}
}

// NetworkInterface 网卡信息
type NetworkInterface struct {
	Name  string   `json:"name"`
	IPs   []string `json:"ips"`
	Mac   string   `json:"mac"`
	Flags []string `json:"flags"`
}

// NetworkInfo 网络信息响应
type NetworkInfo struct {
	Interfaces  []NetworkInterface `json:"interfaces"`
	PublicIP    string             `json:"publicIp"`
	DefaultDNS  *DNSProviderInfo   `json:"defaultDns"`
}

type DNSProviderInfo struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	ProviderType string `json:"providerType"`
}

// GetNetworkInfo 获取网络信息
// GET /api/network/info
func (h *NetworkHandler) GetNetworkInfo(c *gin.Context) {
	var interfaces []NetworkInterface

	// 获取所有网卡
	ifaces, err := net.Interfaces()
	if err == nil {
		for _, iface := range ifaces {
			// 跳过回环接口和未启用的接口
			if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
				continue
			}

			var ips []string
			addrs, err := iface.Addrs()
			if err == nil {
				for _, addr := range addrs {
					// 只获取IPv4地址
					if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
						if ipnet.IP.To4() != nil {
							ips = append(ips, ipnet.IP.String())
						}
					}
				}
			}

			if len(ips) > 0 {
				var flags []string
				if iface.Flags&net.FlagUp != 0 {
					flags = append(flags, "up")
				}
				if iface.Flags&net.FlagBroadcast != 0 {
					flags = append(flags, "broadcast")
				}
				if iface.Flags&net.FlagMulticast != 0 {
					flags = append(flags, "multicast")
				}

				interfaces = append(interfaces, NetworkInterface{
					Name:  iface.Name,
					IPs:   ips,
					Mac:   iface.HardwareAddr.String(),
					Flags: flags,
				})
			}
		}
	}

	// 获取公网IP
	publicIP := h.getPublicIP()

	// 获取默认DNS供应商
	var defaultDNS *DNSProviderInfo
	provider, err := h.dnsService.GetDefault()
	if err == nil && provider != nil {
		defaultDNS = &DNSProviderInfo{
			ID:           provider.ID,
			Name:         provider.Name,
			ProviderType: provider.ProviderType,
		}
	}

	response.Success(c, NetworkInfo{
		Interfaces: interfaces,
		PublicIP:   publicIP,
		DefaultDNS: defaultDNS,
	})
}

// getPublicIP 获取公网IP
func (h *NetworkHandler) getPublicIP() string {
	// 尝试多个公网IP查询服务
	services := []string{
		"https://api.ipify.org?format=text",
		"https://ifconfig.me/ip",
		"https://icanhazip.com",
	}

	for _, url := range services {
		resp, err := http.Get(url)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			continue
		}

		ip := strings.TrimSpace(string(body))
		if ip != "" && net.ParseIP(ip) != nil {
			return ip
		}
	}

	return ""
}

// CreateDNSRecordRequest 创建DNS记录请求
type CreateDNSRecordRequest struct {
	Domain       string `json:"domain" binding:"required"`
	IP           string `json:"ip" binding:"required"`
	RecordType   string `json:"recordType"`   // A, AAAA, CNAME 等
	DnsProviderID uint   `json:"dnsProviderId"` // 可选，不传则使用默认
}

// CreateDNSRecord 创建DNS解析记录
// POST /api/network/dns-record
func (h *NetworkHandler) CreateDNSRecord(c *gin.Context) {
	var req CreateDNSRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if req.RecordType == "" {
		req.RecordType = "A"
	}

	// 获取DNS供应商
	var provider *model.DnsProvider
	var err error

	if req.DnsProviderID > 0 {
		provider, err = h.dnsService.GetByID(req.DnsProviderID)
	} else {
		provider, err = h.dnsService.GetDefault()
	}

	if err != nil || provider == nil {
		response.Error(c, 400, "未找到可用的DNS供应商配置")
		return
	}

	// 创建DNS记录
	if err := h.createDNSRecordForProvider(provider, req.Domain, req.IP, req.RecordType); err != nil {
		response.Error(c, 500, "DNS记录创建失败: "+err.Error())
		return
	}

	response.SuccessWithMessage(c, "DNS记录创建成功", gin.H{
		"domain": req.Domain,
		"ip":     req.IP,
		"type":   req.RecordType,
		"provider": provider.Name,
	})
}

// createDNSRecordForProvider 根据供应商类型创建DNS记录
func (h *NetworkHandler) createDNSRecordForProvider(provider *model.DnsProvider, domain, ip, recordType string) error {
	dnsProvider, err := dnsprovider.NewProvider(dnsprovider.ProviderConfig{
		ProviderType:    provider.ProviderType,
		AccessKeyID:     provider.AccessKeyID,
		AccessKeySecret: provider.AccessKeySecret,
	})
	if err != nil {
		return err
	}

	return dnsProvider.CreateDNSRecord(domain, ip, recordType)
}
