package service

import (
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"nginxops/internal/config"
	"nginxops/internal/model"
	"nginxops/internal/repository"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/dns/alidns"
	"github.com/go-acme/lego/v4/providers/dns/cloudflare"
	"github.com/go-acme/lego/v4/providers/dns/tencentcloud"
	"github.com/go-acme/lego/v4/registration"
)

type AcmeService struct {
	certRepo    *repository.CertificateRepository
	requestRepo *repository.CertificateRequestRepository
	dnsRepo     *repository.DnsProviderRepository
	pending     sync.Map
}

func NewAcmeService() *AcmeService {
	return &AcmeService{
		certRepo:    repository.NewCertificateRepository(),
		requestRepo: repository.NewCertificateRequestRepository(),
		dnsRepo:     repository.NewDnsProviderRepository(),
	}
}

// AcmeUser 实现lego的用户接口
type AcmeUser struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *AcmeUser) GetEmail() string {
	return u.Email
}
func (u AcmeUser) GetRegistration() *registration.Resource {
	return u.Registration
}
func (u *AcmeUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

// RequestCertificateAsync 异步申请证书
func (s *AcmeService) RequestCertificateAsync(certID uint, issuer string, dnsProviderID uint) {
	// 避免重复申请
	if _, loaded := s.pending.LoadOrStore(certID, true); loaded {
		log.Printf("证书 %d 正在申请中，跳过重复请求", certID)
		return
	}

	go func() {
		defer s.pending.Delete(certID)
		s.doRequestCertificate(certID, issuer, dnsProviderID)
	}()
}

// doRequestCertificate 执行证书申请
func (s *AcmeService) doRequestCertificate(certID uint, issuer string, dnsProviderID uint) {
	cert, err := s.certRepo.FindByID(certID)
	if err != nil || cert == nil {
		log.Printf("证书记录不存在: %d", certID)
		return
	}

	// 创建申请记录
	now := time.Now()
	req := &model.CertificateRequest{
		CertificateID: certID,
		Status:        "pending",
		ChallengeType: "dns-01",
		DNSProviderID: dnsProviderID,
		StartedAt:     &now,
	}
	s.requestRepo.Create(req)

	// 更新证书状态
	cert.Status = "pending"
	s.certRepo.Update(cert)

	// 获取DNS服务商配置
	var dnsProvider model.DnsProvider
	if dnsProviderID > 0 {
		dp, err := s.dnsRepo.FindByID(dnsProviderID)
		if err != nil {
			s.updateRequestFailed(req, "DNS服务商配置不存在")
			return
		}
		dnsProvider = *dp
	} else {
		// 获取默认DNS服务商
		providers, _ := s.dnsRepo.FindAll()
		for _, p := range providers {
			if p.IsDefault {
				dnsProvider = p
				break
			}
		}
	}

	// 创建私钥
	privateKey, err := rsa.GenerateKey(nil, 2048)
	if err != nil {
		s.updateRequestFailed(req, "生成私钥失败")
		return
	}

	// 创建用户
	user := &AcmeUser{
		Email: "admin@" + cert.Domain,
		key:   privateKey,
	}

	// 创建配置
	legoConfig := lego.NewConfig(user)
	legoConfig.CADirURL = s.getAcmeUrl(issuer)

	// 创建客户端
	client, err := lego.NewClient(legoConfig)
	if err != nil {
		s.updateRequestFailed(req, "创建ACME客户端失败: "+err.Error())
		return
	}

	// 设置DNS提供商
	s.updateRequestStatus(req, "validating")

	provider, err := s.createDNSProvider(&dnsProvider)
	if err != nil {
		s.updateRequestFailed(req, "配置DNS提供商失败: "+err.Error())
		return
	}

	err = client.Challenge.SetDNS01Provider(provider)
	if err != nil {
		s.updateRequestFailed(req, "设置DNS验证失败: "+err.Error())
		return
	}

	// 注册用户
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		s.updateRequestFailed(req, "注册ACME账户失败: "+err.Error())
		return
	}
	user.Registration = reg

	// 申请证书
	s.updateRequestStatus(req, "issuing")

	domains := s.parseDomains(cert.Domain)
	request := certificate.ObtainRequest{
		Domains: domains,
		Bundle:  true,
	}

	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		s.updateRequestFailed(req, "获取证书失败: "+err.Error())
		cert.Status = "failed"
		s.certRepo.Update(cert)
		return
	}

	// 保存证书文件
	sslPath := config.AppConfig.Nginx.SSLPath
	certFileName := strings.Replace(strings.Replace(cert.Domain, "*.", "wildcard_", -1), ".", "_", -1)
	certFilePath := filepath.Join(sslPath, certFileName+".crt")
	keyFilePath := filepath.Join(sslPath, certFileName+".key")

	if err := os.MkdirAll(sslPath, 0755); err != nil {
		s.updateRequestFailed(req, "创建目录失败")
		return
	}

	// 保存证书
	if err := os.WriteFile(certFilePath, certificates.Certificate, 0644); err != nil {
		s.updateRequestFailed(req, "保存证书失败")
		return
	}

	// 保存私钥
	if err := os.WriteFile(keyFilePath, certificates.PrivateKey, 0600); err != nil {
		s.updateRequestFailed(req, "保存私钥失败")
		return
	}

	// 解析证书过期时间
	expiresAt := time.Now().AddDate(0, 3, 0)
	if certBlock, _ := pem.Decode(certificates.Certificate); certBlock != nil {
		if x509Cert, err := x509.ParseCertificate(certBlock.Bytes); err == nil {
			expiresAt = x509Cert.NotAfter
		}
	}

	// 更新证书记录
	now = time.Now()
	cert.Status = "valid"
	cert.Issuer = s.getIssuerDisplayName(issuer)
	cert.CertPath = certFilePath
	cert.KeyPath = keyFilePath
	cert.IssuedAt = &now
	cert.ExpiresAt = &expiresAt
	s.certRepo.Update(cert)

	// 更新申请记录
	now = time.Now()
	req.Status = "completed"
	req.CompletedAt = &now
	s.requestRepo.Update(req)

	log.Printf("证书申请成功: %s", cert.Domain)
}

func (s *AcmeService) updateRequestStatus(req *model.CertificateRequest, status string) {
	req.Status = status
	s.requestRepo.Update(req)
}

func (s *AcmeService) updateRequestFailed(req *model.CertificateRequest, errorMsg string) {
	now := time.Now()
	req.Status = "failed"
	req.ErrorMessage = errorMsg
	req.CompletedAt = &now
	s.requestRepo.Update(req)
	log.Printf("证书申请失败: %s", errorMsg)
}

func (s *AcmeService) getAcmeUrl(issuer string) string {
	switch strings.ToLower(issuer) {
	case "letsencrypt":
		return "https://acme-v02.api.letsencrypt.org/directory"
	case "letsencrypt-staging":
		return "https://acme-staging-v02.api.letsencrypt.org/directory"
	case "zerossl":
		return "https://acme.zerossl.com/v2/DV90"
	default:
		return "https://acme-v02.api.letsencrypt.org/directory"
	}
}

func (s *AcmeService) getIssuerDisplayName(issuer string) string {
	switch strings.ToLower(issuer) {
	case "letsencrypt":
		return "Let's Encrypt"
	case "letsencrypt-staging":
		return "Let's Encrypt (Staging)"
	case "zerossl":
		return "ZeroSSL"
	default:
		return issuer
	}
}

func (s *AcmeService) parseDomains(domainStr string) []string {
	var domains []string
	for _, domain := range strings.Split(domainStr, ",") {
		trimmed := strings.TrimSpace(domain)
		if trimmed != "" {
			domains = append(domains, trimmed)
		}
	}
	return domains
}

func (s *AcmeService) createDNSProvider(provider *model.DnsProvider) (challenge.Provider, error) {
	switch provider.ProviderType {
	case "aliyun":
		return s.createAliyunProvider(provider)
	case "tencent":
		return s.createTencentProvider(provider)
	case "cloudflare":
		return s.createCloudflareProvider(provider)
	default:
		return nil, fmt.Errorf("不支持的DNS服务商: %s", provider.ProviderType)
	}
}

func (s *AcmeService) createAliyunProvider(provider *model.DnsProvider) (challenge.Provider, error) {
	os.Setenv("ALICLOUD_ACCESS_KEY", provider.AccessKeyID)
	os.Setenv("ALICLOUD_SECRET_KEY", provider.AccessKeySecret)
	return alidns.NewDNSProvider()
}

func (s *AcmeService) createTencentProvider(provider *model.DnsProvider) (challenge.Provider, error) {
	config := tencentcloud.NewDefaultConfig()
	config.SecretID = provider.AccessKeyID
	config.SecretKey = provider.AccessKeySecret
	return tencentcloud.NewDNSProviderConfig(config)
}

func (s *AcmeService) createCloudflareProvider(provider *model.DnsProvider) (challenge.Provider, error) {
	config := cloudflare.NewDefaultConfig()
	config.AuthToken = provider.AccessKeyID
	return cloudflare.NewDNSProviderConfig(config)
}
