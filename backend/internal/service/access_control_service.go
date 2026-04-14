package service

import (
	"fmt"
	"log"
	"nginxops/internal/config"
	"nginxops/internal/model"
	"nginxops/internal/repository"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// AccessControlService 访问控制服务
type AccessControlService struct {
	repo *repository.AccessControlRepository
}

func NewAccessControlService() *AccessControlService {
	return &AccessControlService{
		repo: repository.NewAccessControlRepository(),
	}
}

// ==================== DTO 定义 ====================

type SettingsDto struct {
	GeoEnabled         bool   `json:"geoEnabled"`
	IPBlacklistEnabled bool   `json:"ipBlacklistEnabled"`
	DefaultAction      string `json:"defaultAction"`
}

type IPBlacklistDto struct {
	ID        uint   `json:"id"`
	IPAddress string `json:"ipAddress"`
	Note      string `json:"note"`
	Enabled   bool   `json:"enabled"`
}

type GeoRuleDto struct {
	ID          uint   `json:"id"`
	CountryCode string `json:"countryCode"`
	Action      string `json:"action"`
	Note        string `json:"note"`
	Enabled     bool   `json:"enabled"`
}

type SiteIPBlacklistDto struct {
	ID        uint   `json:"id"`
	SiteID    uint   `json:"siteId"`
	IPAddress string `json:"ipAddress"`
	Note      string `json:"note"`
	Enabled   bool   `json:"enabled"`
}

type SiteGeoRuleDto struct {
	ID          uint   `json:"id"`
	SiteID      uint   `json:"siteId"`
	CountryCode string `json:"countryCode"`
	Action      string `json:"action"`
	Note        string `json:"note"`
	Enabled     bool   `json:"enabled"`
}

// ==================== 全局设置 ====================

func (s *AccessControlService) GetSettings() (*SettingsDto, error) {
	settings, err := s.repo.GetSettings()
	if err != nil {
		return nil, err
	}
	return &SettingsDto{
		GeoEnabled:         settings.GeoEnabled,
		IPBlacklistEnabled: settings.IPBlacklistEnabled,
		DefaultAction:      settings.DefaultAction,
	}, nil
}

func (s *AccessControlService) UpdateSettings(dto *SettingsDto) error {
	settings, err := s.repo.GetSettings()
	if err != nil {
		return err
	}
	settings.GeoEnabled = dto.GeoEnabled
	settings.IPBlacklistEnabled = dto.IPBlacklistEnabled
	settings.DefaultAction = dto.DefaultAction

	if err := s.repo.UpdateSettings(settings); err != nil {
		return err
	}

	// 同步配置到 Nginx
	return s.SyncConfig()
}

// ==================== 全局 IP 黑名单 ====================

func (s *AccessControlService) ListIPBlacklist() ([]model.IPBlacklist, error) {
	return s.repo.ListIPBlacklist()
}

func (s *AccessControlService) CreateIPBlacklist(dto *IPBlacklistDto) (*model.IPBlacklist, error) {
	// 验证 IP 格式
	if !isValidIPorCIDR(dto.IPAddress) {
		return nil, fmt.Errorf("无效的 IP 地址或 CIDR 格式: %s", dto.IPAddress)
	}

	item := &model.IPBlacklist{
		IPAddress: dto.IPAddress,
		Note:      dto.Note,
		Enabled:   dto.Enabled,
	}
	if err := s.repo.CreateIPBlacklist(item); err != nil {
		return nil, err
	}

	// 同步配置
	_ = s.SyncConfig()
	return item, nil
}

func (s *AccessControlService) UpdateIPBlacklist(id uint, dto *IPBlacklistDto) (*model.IPBlacklist, error) {
	item, err := s.repo.FindIPBlacklistByID(id)
	if err != nil {
		return nil, fmt.Errorf("IP 黑名单项不存在")
	}

	if dto.IPAddress != "" {
		if !isValidIPorCIDR(dto.IPAddress) {
			return nil, fmt.Errorf("无效的 IP 地址或 CIDR 格式: %s", dto.IPAddress)
		}
		item.IPAddress = dto.IPAddress
	}
	item.Note = dto.Note
	item.Enabled = dto.Enabled

	if err := s.repo.UpdateIPBlacklist(item); err != nil {
		return nil, err
	}

	_ = s.SyncConfig()
	return item, nil
}

func (s *AccessControlService) DeleteIPBlacklist(id uint) error {
	if err := s.repo.DeleteIPBlacklist(id); err != nil {
		return err
	}
	_ = s.SyncConfig()
	return nil
}

func (s *AccessControlService) ToggleIPBlacklist(id uint, enabled bool) error {
	item, err := s.repo.FindIPBlacklistByID(id)
	if err != nil {
		return fmt.Errorf("IP 黑名单项不存在")
	}
	item.Enabled = enabled
	if err := s.repo.UpdateIPBlacklist(item); err != nil {
		return err
	}
	_ = s.SyncConfig()
	return nil
}

// ==================== 全局 Geo 规则 ====================

func (s *AccessControlService) ListGeoRules() ([]model.GeoRule, error) {
	return s.repo.ListGeoRules()
}

func (s *AccessControlService) CreateGeoRule(dto *GeoRuleDto) (*model.GeoRule, error) {
	// 验证国家代码
	if len(dto.CountryCode) < 2 || len(dto.CountryCode) > 3 {
		return nil, fmt.Errorf("无效的国家代码: %s", dto.CountryCode)
	}
	// 验证动作
	if dto.Action != "allow" && dto.Action != "block" {
		dto.Action = "block"
	}

	item := &model.GeoRule{
		CountryCode: strings.ToUpper(dto.CountryCode),
		Action:      dto.Action,
		Note:        dto.Note,
		Enabled:     dto.Enabled,
	}
	if err := s.repo.CreateGeoRule(item); err != nil {
		return nil, err
	}

	_ = s.SyncConfig()
	return item, nil
}

func (s *AccessControlService) UpdateGeoRule(id uint, dto *GeoRuleDto) (*model.GeoRule, error) {
	item, err := s.repo.FindGeoRuleByID(id)
	if err != nil {
		return nil, fmt.Errorf("Geo 规则不存在")
	}

	if dto.CountryCode != "" {
		if len(dto.CountryCode) < 2 || len(dto.CountryCode) > 3 {
			return nil, fmt.Errorf("无效的国家代码: %s", dto.CountryCode)
		}
		item.CountryCode = strings.ToUpper(dto.CountryCode)
	}
	if dto.Action != "" {
		if dto.Action != "allow" && dto.Action != "block" {
			dto.Action = "block"
		}
		item.Action = dto.Action
	}
	item.Note = dto.Note
	item.Enabled = dto.Enabled

	if err := s.repo.UpdateGeoRule(item); err != nil {
		return nil, err
	}

	_ = s.SyncConfig()
	return item, nil
}

func (s *AccessControlService) DeleteGeoRule(id uint) error {
	if err := s.repo.DeleteGeoRule(id); err != nil {
		return err
	}
	_ = s.SyncConfig()
	return nil
}

func (s *AccessControlService) ToggleGeoRule(id uint, enabled bool) error {
	item, err := s.repo.FindGeoRuleByID(id)
	if err != nil {
		return fmt.Errorf("Geo 规则不存在")
	}
	item.Enabled = enabled
	if err := s.repo.UpdateGeoRule(item); err != nil {
		return err
	}
	_ = s.SyncConfig()
	return nil
}

// ==================== 站点专属 IP 黑名单 ====================

func (s *AccessControlService) ListSiteIPBlacklist(siteID uint) ([]model.SiteIPBlacklist, error) {
	return s.repo.ListSiteIPBlacklist(siteID)
}

func (s *AccessControlService) CreateSiteIPBlacklist(dto *SiteIPBlacklistDto) (*model.SiteIPBlacklist, error) {
	if !isValidIPorCIDR(dto.IPAddress) {
		return nil, fmt.Errorf("无效的 IP 地址或 CIDR 格式: %s", dto.IPAddress)
	}

	item := &model.SiteIPBlacklist{
		SiteID:    dto.SiteID,
		IPAddress: dto.IPAddress,
		Note:      dto.Note,
		Enabled:   dto.Enabled,
	}
	if err := s.repo.CreateSiteIPBlacklist(item); err != nil {
		return nil, err
	}

	_ = s.SyncConfig()
	return item, nil
}

func (s *AccessControlService) DeleteSiteIPBlacklist(id uint) error {
	if err := s.repo.DeleteSiteIPBlacklist(id); err != nil {
		return err
	}
	_ = s.SyncConfig()
	return nil
}

// ==================== 站点专属 Geo 规则 ====================

func (s *AccessControlService) ListSiteGeoRules(siteID uint) ([]model.SiteGeoRule, error) {
	return s.repo.ListSiteGeoRules(siteID)
}

func (s *AccessControlService) CreateSiteGeoRule(dto *SiteGeoRuleDto) (*model.SiteGeoRule, error) {
	if len(dto.CountryCode) < 2 || len(dto.CountryCode) > 3 {
		return nil, fmt.Errorf("无效的国家代码: %s", dto.CountryCode)
	}
	if dto.Action != "allow" && dto.Action != "block" {
		dto.Action = "block"
	}

	item := &model.SiteGeoRule{
		SiteID:      dto.SiteID,
		CountryCode: strings.ToUpper(dto.CountryCode),
		Action:      dto.Action,
		Note:        dto.Note,
		Enabled:     dto.Enabled,
	}
	if err := s.repo.CreateSiteGeoRule(item); err != nil {
		return nil, err
	}

	_ = s.SyncConfig()
	return item, nil
}

func (s *AccessControlService) DeleteSiteGeoRule(id uint) error {
	if err := s.repo.DeleteSiteGeoRule(id); err != nil {
		return err
	}
	_ = s.SyncConfig()
	return nil
}

// ==================== 配置同步 ====================

func (s *AccessControlService) SyncConfig() error {
	settings, err := s.repo.GetSettings()
	if err != nil {
		return err
	}

	confDir := filepath.Join(config.AppConfig.Nginx.ConfDir, "access-control")
	if err := os.MkdirAll(confDir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %v", err)
	}

	// 生成全局 IP 黑名单配置
	ipBlacklist, err := s.repo.FindEnabledIPBlacklist()
	if err != nil {
		return err
	}
	if err := s.generateIPBlacklistConfig(confDir, ipBlacklist, settings.IPBlacklistEnabled); err != nil {
		return err
	}

	// 生成全局 Geo 规则配置
	geoRules, err := s.repo.FindEnabledGeoRules()
	if err != nil {
		return err
	}
	if err := s.generateGeoConfig(confDir, geoRules, settings); err != nil {
		return err
	}

	// 重新加载 Nginx
	s.reloadNginx()
	return nil
}

func (s *AccessControlService) generateIPBlacklistConfig(confDir string, list []model.IPBlacklist, enabled bool) error {
	var sb strings.Builder

	sb.WriteString("# 自动生成的 IP 黑名单配置 - 请勿手动修改\n")
	sb.WriteString(fmt.Sprintf("# 生成时间: 略\n\n"))

	if enabled && len(list) > 0 {
		sb.WriteString("geo $global_blocked_ip {\n")
		sb.WriteString("    default 0;\n")
		for _, item := range list {
			sb.WriteString(fmt.Sprintf("    %s 1;\n", item.IPAddress))
		}
		sb.WriteString("}\n")
	} else {
		// 禁用状态，生成空配置
		sb.WriteString("geo $global_blocked_ip {\n")
		sb.WriteString("    default 0;\n")
		sb.WriteString("}\n")
	}

	confPath := filepath.Join(confDir, "global_ip_blacklist.conf")
	return os.WriteFile(confPath, []byte(sb.String()), 0644)
}

func (s *AccessControlService) generateGeoConfig(confDir string, rules []model.GeoRule, settings *model.AccessControlSettings) error {
	var sb strings.Builder

	sb.WriteString("# 自动生成的 Geo 封锁配置 - 请勿手动修改\n\n")

	if settings.GeoEnabled && len(rules) > 0 {
		sb.WriteString("map $geoip_country_code $global_geo_action {\n")
		sb.WriteString(fmt.Sprintf("    default %s;\n", settings.DefaultAction))
		for _, rule := range rules {
			sb.WriteString(fmt.Sprintf("    %s %s;\n", rule.CountryCode, rule.Action))
		}
		sb.WriteString("}\n")
	} else {
		// 禁用状态，使用固定值映射（不依赖 GeoIP 模块）
		// 使用空字符串作为 key，default 会匹配所有情况
		sb.WriteString("map $remote_addr $global_geo_action {\n")
		sb.WriteString("    default allow;\n")
		sb.WriteString("}\n")
	}

	confPath := filepath.Join(confDir, "global_geo.conf")
	return os.WriteFile(confPath, []byte(sb.String()), 0644)
}

func (s *AccessControlService) reloadNginx() {
	testCmd := config.AppConfig.Nginx.TestCommand
	reloadCmd := config.AppConfig.Nginx.ReloadCommand

	test := exec.Command("sh", "-c", testCmd)
	if err := test.Run(); err != nil {
		log.Printf("Nginx 配置测试失败: %v", err)
		return
	}

	reload := exec.Command("sh", "-c", reloadCmd)
	if err := reload.Run(); err != nil {
		log.Printf("Nginx 重载失败: %v", err)
	} else {
		log.Println("访问控制配置已同步到 Nginx")
	}
}

// ==================== 站点配置生成辅助 ====================

// GetSiteAccessControlConfig 生成站点访问控制配置片段
// 注意：geo/map 指令必须在 http 块中，所以这里只生成条件判断逻辑
func (s *AccessControlService) GetSiteAccessControlConfig(siteID uint, mode string) (string, error) {
	var sb strings.Builder

	// 获取全局设置
	settings, err := s.repo.GetSettings()
	if err != nil {
		return "", err
	}

	// 先同步站点专属配置到独立文件（这些文件会在 http 块中被 include）
	if err := s.syncSiteSpecificConfig(siteID, mode, settings); err != nil {
		log.Printf("同步站点专属访问控制配置失败: %v", err)
	}

	switch mode {
	case "inherit":
		// 继承全局规则
		sb.WriteString("    # 访问控制 - 继承全局规则\n")
		if settings.GeoEnabled {
			sb.WriteString("    if ($global_geo_action = block) { return 403; }\n")
		}
		if settings.IPBlacklistEnabled {
			sb.WriteString("    if ($global_blocked_ip) { return 403; }\n")
		}

	case "merge":
		// 合并全局 + 站点规则
		sb.WriteString("    # 访问控制 - 合并模式\n")
		if settings.GeoEnabled {
			sb.WriteString("    if ($global_geo_action = block) { return 403; }\n")
		}
		if settings.IPBlacklistEnabled {
			sb.WriteString("    if ($global_blocked_ip) { return 403; }\n")
		}
		// 站点专属规则
		sb.WriteString(s.generateSiteSpecificConditions(siteID))

	case "override":
		// 仅使用站点规则
		sb.WriteString("    # 访问控制 - 站点专属规则\n")
		sb.WriteString(s.generateSiteSpecificConditions(siteID))
	}

	return sb.String(), nil
}

// syncSiteSpecificConfig 同步站点专属配置到独立文件
func (s *AccessControlService) syncSiteSpecificConfig(siteID uint, mode string, settings *model.AccessControlSettings) error {
	confDir := filepath.Join(config.AppConfig.Nginx.ConfDir, "access-control")
	if err := os.MkdirAll(confDir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %v", err)
	}

	// 在 override 模式下不使用站点专属配置文件，使用内联变量
	// 但我们仍然需要在 http 块中定义变量

	// 生成站点 IP 黑名单配置（geo 指令必须在 http 块中）
	ipList, err := s.repo.FindEnabledSiteIPBlacklist(siteID)
	if err == nil && len(ipList) > 0 && (mode == "merge" || mode == "override") {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("# 站点 %d IP 黑名单配置\n", siteID))
		sb.WriteString(fmt.Sprintf("geo $site_%d_blocked_ip {\n", siteID))
		sb.WriteString("    default 0;\n")
		for _, item := range ipList {
			sb.WriteString(fmt.Sprintf("    %s 1;\n", item.IPAddress))
		}
		sb.WriteString("}\n")
		confPath := filepath.Join(confDir, fmt.Sprintf("site_%d_ip_blacklist.conf", siteID))
		if err := os.WriteFile(confPath, []byte(sb.String()), 0644); err != nil {
			log.Printf("写入站点 IP 黑名单配置失败: %v", err)
		}
	} else {
		// 清理可能存在的旧配置
		confPath := filepath.Join(confDir, fmt.Sprintf("site_%d_ip_blacklist.conf", siteID))
		os.Remove(confPath)
	}

	// 生成站点 Geo 规则配置（map 指令必须在 http 块中）
	geoRules, err := s.repo.FindEnabledSiteGeoRules(siteID)
	if err == nil && len(geoRules) > 0 && (mode == "merge" || mode == "override") {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("# 站点 %d Geo 规则配置\n", siteID))
		sb.WriteString(fmt.Sprintf("map $geoip_country_code $site_%d_geo_action {\n", siteID))
		sb.WriteString("    default allow;\n")
		for _, rule := range geoRules {
			sb.WriteString(fmt.Sprintf("    %s %s;\n", rule.CountryCode, rule.Action))
		}
		sb.WriteString("}\n")
		confPath := filepath.Join(confDir, fmt.Sprintf("site_%d_geo.conf", siteID))
		if err := os.WriteFile(confPath, []byte(sb.String()), 0644); err != nil {
			log.Printf("写入站点 Geo 规则配置失败: %v", err)
		}
	} else {
		// 清理可能存在的旧配置
		confPath := filepath.Join(confDir, fmt.Sprintf("site_%d_geo.conf", siteID))
		os.Remove(confPath)
	}

	return nil
}

// generateSiteSpecificConditions 生成站点专属的条件判断逻辑
func (s *AccessControlService) generateSiteSpecificConditions(siteID uint) string {
	var sb strings.Builder

	// 站点 IP 黑名单条件
	ipList, err := s.repo.FindEnabledSiteIPBlacklist(siteID)
	if err == nil && len(ipList) > 0 {
		sb.WriteString(fmt.Sprintf("    if ($site_%d_blocked_ip) { return 403; }\n", siteID))
	}

	// 站点 Geo 规则条件
	geoRules, err := s.repo.FindEnabledSiteGeoRules(siteID)
	if err == nil && len(geoRules) > 0 {
		sb.WriteString(fmt.Sprintf("    if ($site_%d_geo_action = block) { return 403; }\n", siteID))
	}

	return sb.String()
}

// ==================== 辅助函数 ====================

func isValidIPorCIDR(s string) bool {
	// 简单验证：包含数字和点，可能有斜杠
	if len(s) == 0 || len(s) > 50 {
		return false
	}
	for _, c := range s {
		if !((c >= '0' && c <= '9') || c == '.' || c == '/' || c == ':' ||
			(c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}
