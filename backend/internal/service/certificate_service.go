package service

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"nginxops/internal/config"
	"nginxops/internal/model"
	"nginxops/internal/repository"
	"os"
	"strings"
	"time"
)

type CertificateService struct {
	repo *repository.CertificateRepository
}

func NewCertificateService() *CertificateService {
	return &CertificateService{
		repo: repository.NewCertificateRepository(),
	}
}

type CertificateImportDto struct {
	Domain          string `json:"domain"`
	Certificate     string `json:"certificate"`
	PrivateKey      string `json:"privateKey"`
	CertificateChain string `json:"certificateChain"`
}

func (s *CertificateService) ListAll() ([]model.Certificate, error) {
	return s.repo.FindAll()
}

func (s *CertificateService) ListAvailable() ([]model.Certificate, error) {
	return s.repo.FindByStatus("valid")
}

func (s *CertificateService) List(page, size int, status string) ([]model.Certificate, int64, error) {
	return s.repo.FindPage(page, size, status)
}

func (s *CertificateService) GetByID(id uint) (*model.Certificate, error) {
	return s.repo.FindByID(id)
}

func (s *CertificateService) Create(cert *model.Certificate) (*model.Certificate, error) {
	cert.Status = "pending"
	cert.AutoRenew = true
	if err := s.repo.Create(cert); err != nil {
		return nil, err
	}
	return cert, nil
}

func (s *CertificateService) Update(id uint, cert *model.Certificate) (*model.Certificate, error) {
	existing, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("证书不存在")
	}

	existing.Domain = cert.Domain
	existing.Issuer = cert.Issuer
	existing.AutoRenew = cert.AutoRenew
	if err := s.repo.Update(existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *CertificateService) Delete(id uint) error {
	return s.repo.Delete(id)
}

func (s *CertificateService) Renew(id uint) (*model.Certificate, error) {
	cert, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("证书不存在")
	}

	now := time.Now()
	expiresAt := now.AddDate(0, 3, 0)
	cert.Status = "valid"
	cert.IssuedAt = &now
	cert.ExpiresAt = &expiresAt
	if err := s.repo.Update(cert); err != nil {
		return nil, err
	}
	return cert, nil
}

func (s *CertificateService) RequestCertificate(id uint, issuer string, dnsProviderID uint) error {
	cert, err := s.repo.FindByID(id)
	if err != nil {
		return fmt.Errorf("证书不存在")
	}

	cert.Status = "pending"
	if err := s.repo.Update(cert); err != nil {
		return err
	}

	// 异步申请证书（ACME）
	acmeService := NewAcmeService()
	acmeService.RequestCertificateAsync(id, issuer, dnsProviderID)
	log.Printf("证书申请已提交: %s, issuer: %s, dnsProviderId: %d", cert.Domain, issuer, dnsProviderID)
	return nil
}

func (s *CertificateService) ToggleAutoRenew(id uint, autoRenew bool) error {
	cert, err := s.repo.FindByID(id)
	if err != nil {
		return fmt.Errorf("证书不存在")
	}

	cert.AutoRenew = autoRenew
	return s.repo.Update(cert)
}

func (s *CertificateService) ImportCertificate(dto *CertificateImportDto) (*model.Certificate, error) {
	if dto.Certificate == "" {
		return nil, fmt.Errorf("证书内容不能为空")
	}
	if dto.PrivateKey == "" {
		return nil, fmt.Errorf("私钥内容不能为空")
	}

	// 解析证书
	x509Cert, err := s.parseCertificate(dto.Certificate)
	if err != nil {
		return nil, err
	}

	// 获取域名
	domain := dto.Domain
	if domain == "" {
		domain = s.extractDomainFromCert(x509Cert)
	}
	if domain == "" {
		return nil, fmt.Errorf("无法从证书中解析域名，请手动指定域名")
	}

	// 检查是否已存在
	existing, _ := s.repo.FindByDomain(domain)

	// 准备证书文件路径
	sslPath := config.AppConfig.Nginx.SSLPath
	certFileName := strings.Replace(strings.Replace(domain, "*.", "wildcard_", -1), ".", "_", -1)
	certPath := fmt.Sprintf("%s/%s.crt", sslPath, certFileName)
	keyPath := fmt.Sprintf("%s/%s.key", sslPath, certFileName)

	// 写入证书文件
	if err := s.writeCertificateFile(certPath, dto.Certificate, dto.CertificateChain); err != nil {
		return nil, fmt.Errorf("写入证书文件失败: %v", err)
	}
	if err := s.writePrivateKeyFile(keyPath, dto.PrivateKey); err != nil {
		return nil, fmt.Errorf("写入私钥文件失败: %v", err)
	}

	// 确定状态
	status := "valid"
	if x509Cert.NotAfter.Before(time.Now()) {
		status = "expired"
	}

	issuerName := s.getIssuerName(x509Cert)
	issuedAt := x509Cert.NotBefore
	expiresAt := x509Cert.NotAfter

	// 创建或更新证书记录
	var cert *model.Certificate
	if existing != nil {
		cert = existing
		cert.CertPath = certPath
		cert.KeyPath = keyPath
		cert.Issuer = issuerName
		cert.IssuedAt = &issuedAt
		cert.ExpiresAt = &expiresAt
		cert.Status = status
		if err := s.repo.Update(cert); err != nil {
			return nil, err
		}
	} else {
		cert = &model.Certificate{
			Domain:    domain,
			CertPath:  certPath,
			KeyPath:   keyPath,
			Issuer:    issuerName,
			IssuedAt:  &issuedAt,
			ExpiresAt: &expiresAt,
			Status:    status,
			AutoRenew: false,
		}
		if err := s.repo.Create(cert); err != nil {
			return nil, err
		}
	}

	return cert, nil
}

func (s *CertificateService) parseCertificate(pemContent string) (*x509.Certificate, error) {
	block, _ := pem.Decode([]byte(pemContent))
	if block == nil {
		return nil, fmt.Errorf("无效的证书格式，无法找到BEGIN CERTIFICATE标记")
	}

	return x509.ParseCertificate(block.Bytes)
}

func (s *CertificateService) extractDomainFromCert(cert *x509.Certificate) string {
	// 尝试从CN中获取
	if cert.Subject.CommonName != "" {
		return cert.Subject.CommonName
	}

	// 尝试从SAN中获取
	for _, name := range cert.DNSNames {
		return name
	}

	return ""
}

func (s *CertificateService) getIssuerName(cert *x509.Certificate) string {
	issuer := cert.Issuer.CommonName

	// 常见CA识别
	issuerOrg := strings.Join(cert.Issuer.Organization, " ")
	if strings.Contains(issuerOrg, "Let's Encrypt") {
		return "Let's Encrypt"
	} else if strings.Contains(issuerOrg, "ZeroSSL") {
		return "ZeroSSL"
	} else if strings.Contains(issuerOrg, "DigiCert") {
		return "DigiCert"
	} else if strings.Contains(issuerOrg, "GlobalSign") {
		return "GlobalSign"
	} else if strings.Contains(issuerOrg, "Comodo") || strings.Contains(issuerOrg, "Sectigo") {
		return "Sectigo"
	}

	if issuer != "" {
		return issuer
	}

	return "Unknown"
}

func (s *CertificateService) writeCertificateFile(path, certificate, chain string) error {
	if err := os.MkdirAll(strings.Replace(path, "/"+strings.Split(path, "/")[len(strings.Split(path, "/"))-1], "", 1), 0755); err != nil {
		return err
	}

	content := strings.TrimSpace(certificate) + "\n"
	if chain != "" {
		content += strings.TrimSpace(chain) + "\n"
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return err
	}
	log.Printf("证书文件已写入: %s", path)
	return nil
}

func (s *CertificateService) writePrivateKeyFile(path, privateKey string) error {
	if err := os.MkdirAll(strings.Replace(path, "/"+strings.Split(path, "/")[len(strings.Split(path, "/"))-1], "", 1), 0755); err != nil {
		return err
	}

	if err := os.WriteFile(path, []byte(strings.TrimSpace(privateKey)+"\n"), 0600); err != nil {
		return err
	}
	log.Printf("私钥文件已写入: %s", path)
	return nil
}
