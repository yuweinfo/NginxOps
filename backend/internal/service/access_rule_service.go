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

// AccessRuleService 访问规则服务
type AccessRuleService struct {
	repo *repository.AccessRuleRepository
}

func NewAccessRuleService() *AccessRuleService {
	return &AccessRuleService{
		repo: repository.NewAccessRuleRepository(),
	}
}

// ==================== DTO 定义 ====================

type AccessRuleDto struct {
	ID          uint                `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Enabled     bool                `json:"enabled"`
	Items       []AccessRuleItemDto `json:"items"`
}

type AccessRuleItemDto struct {
	ID          uint   `json:"id"`
	RuleID      uint   `json:"ruleId"`
	ItemType    string `json:"itemType"`    // ip / geo
	IPAddress   string `json:"ipAddress"`   // 当 ItemType=ip 时使用
	CountryCode string `json:"countryCode"` // 当 ItemType=geo 时使用
	Action      string `json:"action"`      // 当 ItemType=geo 时使用: allow/block
	Note        string `json:"note"`
}

type AccessRuleSummaryDto struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
	IPCount     int    `json:"ipCount"`
	GeoCount    int    `json:"geoCount"`
	SiteCount   int64  `json:"siteCount"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type SetSiteRulesDto struct {
	RuleIDs []uint `json:"ruleIds"`
}

// ==================== 规则 CRUD ====================

func (s *AccessRuleService) ListRules() ([]AccessRuleSummaryDto, error) {
	rules, err := s.repo.ListRules()
	if err != nil {
		return nil, err
	}

	siteCounts, err := s.repo.GetRulesSiteCounts()
	if err != nil {
		log.Printf("获取规则引用数失败: %v", err)
		siteCounts = make(map[uint]int64)
	}

	result := make([]AccessRuleSummaryDto, len(rules))
	for i, rule := range rules {
		ipCount, geoCount := countRuleItems(rule.Items)
		result[i] = AccessRuleSummaryDto{
			ID:          rule.ID,
			Name:        rule.Name,
			Description: rule.Description,
			Enabled:     rule.Enabled,
			IPCount:     ipCount,
			GeoCount:    geoCount,
			SiteCount:   siteCounts[rule.ID],
			CreatedAt:   rule.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   rule.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}
	return result, nil
}

func (s *AccessRuleService) GetRule(id uint) (*AccessRuleDto, error) {
	rule, err := s.repo.FindRuleByID(id)
	if err != nil {
		return nil, fmt.Errorf("规则不存在")
	}
	return ruleToDto(rule), nil
}

func (s *AccessRuleService) CreateRule(dto *AccessRuleDto) (*model.AccessRule, error) {
	if dto.Name == "" {
		return nil, fmt.Errorf("规则名称不能为空")
	}

	rule := &model.AccessRule{
		Name:        dto.Name,
		Description: dto.Description,
		Enabled:     dto.Enabled,
	}

	// 构建条目
	for _, itemDto := range dto.Items {
		if err := validateItemDto(&itemDto); err != nil {
			return nil, err
		}
		item := dtoToItem(itemDto)
		rule.Items = append(rule.Items, item)
	}

	if err := s.repo.CreateRule(rule); err != nil {
		return nil, err
	}

	_ = s.SyncAllConfigs()
	return rule, nil
}

func (s *AccessRuleService) UpdateRule(id uint, dto *AccessRuleDto) (*model.AccessRule, error) {
	rule, err := s.repo.FindRuleByID(id)
	if err != nil {
		return nil, fmt.Errorf("规则不存在")
	}

	if dto.Name != "" {
		rule.Name = dto.Name
	}
	rule.Description = dto.Description
	rule.Enabled = dto.Enabled

	// 替换所有条目：先删后增
	if err := s.repo.DeleteItemsByRuleID(id); err != nil {
		return nil, err
	}

	rule.Items = nil
	for _, itemDto := range dto.Items {
		if err := validateItemDto(&itemDto); err != nil {
			return nil, err
		}
		item := dtoToItem(itemDto)
		item.RuleID = id
		item.ID = 0 // 确保新建
		rule.Items = append(rule.Items, item)
	}

	if err := s.repo.UpdateRule(rule); err != nil {
		return nil, err
	}

	_ = s.SyncAllConfigs()
	return rule, nil
}

func (s *AccessRuleService) DeleteRule(id uint) error {
	// 检查是否有站点引用
	count, err := s.repo.GetRuleSiteCount(id)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("该规则正在被 %d 个站点使用，无法删除", count)
	}

	if err := s.repo.DeleteRule(id); err != nil {
		return err
	}
	_ = s.SyncAllConfigs()
	return nil
}

func (s *AccessRuleService) ToggleRule(id uint, enabled bool) error {
	rule, err := s.repo.FindRuleByID(id)
	if err != nil {
		return fmt.Errorf("规则不存在")
	}
	rule.Enabled = enabled
	if err := s.repo.UpdateRule(rule); err != nil {
		return err
	}
	_ = s.SyncAllConfigs()
	return nil
}

// ==================== 站点-规则关联 ====================

func (s *AccessRuleService) GetSiteRuleIDs(siteID uint) ([]uint, error) {
	return s.repo.GetSiteRuleIDs(siteID)
}

func (s *AccessRuleService) SetSiteRules(siteID uint, dto *SetSiteRulesDto) error {
	// 验证所有规则ID存在且启用
	for _, ruleID := range dto.RuleIDs {
		rule, err := s.repo.FindRuleByID(ruleID)
		if err != nil {
			return fmt.Errorf("规则 %d 不存在", ruleID)
		}
		if !rule.Enabled {
			return fmt.Errorf("规则 '%s' 已禁用，无法关联", rule.Name)
		}
	}

	if err := s.repo.SetSiteRules(siteID, dto.RuleIDs); err != nil {
		return err
	}
	_ = s.SyncAllConfigs()
	return nil
}

func (s *AccessRuleService) GetRulesForSite(siteID uint) ([]AccessRuleSummaryDto, error) {
	rules, err := s.repo.GetRulesBySiteID(siteID)
	if err != nil {
		return nil, err
	}

	result := make([]AccessRuleSummaryDto, len(rules))
	for i, rule := range rules {
		ipCount, geoCount := countRuleItems(rule.Items)
		result[i] = AccessRuleSummaryDto{
			ID:          rule.ID,
			Name:        rule.Name,
			Description: rule.Description,
			Enabled:     rule.Enabled,
			IPCount:     ipCount,
			GeoCount:    geoCount,
		}
	}
	return result, nil
}

// ==================== 配置同步与规则合并 ====================

// MergedAccessRules 合并后的访问控制规则
type MergedAccessRules struct {
	IPList     []string           // 合并后的 IP 黑名单列表（去重）
	GeoRules   map[string]string  // 合并后的 Geo 规则（countryCode -> action）
}

// MergeRules 合并多条规则为一个统一的规则集
// 逻辑：
// - IP黑名单：取并集（去重）
// - Geo规则：相同国家代码，如果任一规则 block 则 block；如果所有规则都 allow 则 allow
//   优先级：block > allow（安全优先）
func (s *AccessRuleService) MergeRules(rules []model.AccessRule) *MergedAccessRules {
	merged := &MergedAccessRules{
		IPList:   []string{},
		GeoRules: make(map[string]string),
	}

	ipSet := make(map[string]bool)

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}
		for _, item := range rule.Items {
			switch item.ItemType {
			case "ip":
				if item.IPAddress != "" {
					ipSet[item.IPAddress] = true
				}
			case "geo":
				if item.CountryCode != "" {
					cc := strings.ToUpper(item.CountryCode)
					existing, exists := merged.GeoRules[cc]
					if !exists {
						merged.GeoRules[cc] = item.Action
					} else {
						// 安全优先：block 覆盖 allow
						if existing == "allow" && item.Action == "block" {
							merged.GeoRules[cc] = "block"
						}
					}
				}
			}
		}
	}

	for ip := range ipSet {
		merged.IPList = append(merged.IPList, ip)
	}

	return merged
}

// GetSiteMergedRules 获取站点的合并规则
func (s *AccessRuleService) GetSiteMergedRules(siteID uint) (*MergedAccessRules, error) {
	rules, err := s.repo.GetRulesBySiteID(siteID)
	if err != nil {
		return nil, err
	}
	return s.MergeRules(rules), nil
}

// SyncAllConfigs 同步所有访问控制配置到 Nginx
func (s *AccessRuleService) SyncAllConfigs() error {
	confDir := filepath.Join(config.AppConfig.Nginx.ConfDir, "access-control")
	if err := os.MkdirAll(confDir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %v", err)
	}

	// 生成全局合并配置（所有启用的规则合并）
	rules, err := s.repo.FindEnabledRules()
	if err != nil {
		return err
	}
	merged := s.MergeRules(rules)

	// 生成全局 IP 黑名单配置
	if err := s.generateIPBlacklistConfig(confDir, merged.IPList); err != nil {
		return err
	}

	// 生成全局 Geo 规则配置
	if err := s.generateGeoConfig(confDir, merged.GeoRules); err != nil {
		return err
	}

	// 生成各站点的专属配置
	if err := s.syncAllSiteConfigs(confDir); err != nil {
		log.Printf("同步站点配置失败: %v", err)
	}

	s.reloadNginx()
	return nil
}

// GetSiteAccessControlConfig 生成站点访问控制配置片段
// 在新的规则模式下，站点通过关联规则来控制访问
func (s *AccessRuleService) GetSiteAccessControlConfig(siteID uint) (string, error) {
	rules, err := s.repo.GetRulesBySiteID(siteID)
	if err != nil {
		return "", err
	}

	if len(rules) == 0 {
		return "", nil
	}

	merged := s.MergeRules(rules)
	return s.generateSiteConditions(siteID, merged), nil
}

// ==================== 配置生成 ====================

func (s *AccessRuleService) generateIPBlacklistConfig(confDir string, ipList []string) error {
	var sb strings.Builder
	sb.WriteString("# 自动生成的 IP 黑名单配置 - 请勿手动修改\n\n")

	if len(ipList) > 0 {
		sb.WriteString("geo $global_blocked_ip {\n")
		sb.WriteString("    default 0;\n")
		for _, ip := range ipList {
			sb.WriteString(fmt.Sprintf("    %s 1;\n", ip))
		}
		sb.WriteString("}\n")
	} else {
		sb.WriteString("geo $global_blocked_ip {\n")
		sb.WriteString("    default 0;\n")
		sb.WriteString("}\n")
	}

	confPath := filepath.Join(confDir, "global_ip_blacklist.conf")
	return os.WriteFile(confPath, []byte(sb.String()), 0644)
}

func (s *AccessRuleService) generateGeoConfig(confDir string, geoRules map[string]string) error {
	var sb strings.Builder
	sb.WriteString("# 自动生成的 Geo 封锁配置 - 请勿手动修改\n\n")

	if len(geoRules) > 0 {
		sb.WriteString("map $geoip_country_code $global_geo_action {\n")
		sb.WriteString("    default allow;\n")
		for cc, action := range geoRules {
			sb.WriteString(fmt.Sprintf("    %s %s;\n", cc, action))
		}
		sb.WriteString("}\n")
	} else {
		sb.WriteString("map $remote_addr $global_geo_action {\n")
		sb.WriteString("    default allow;\n")
		sb.WriteString("}\n")
	}

	confPath := filepath.Join(confDir, "global_geo.conf")
	return os.WriteFile(confPath, []byte(sb.String()), 0644)
}

func (s *AccessRuleService) syncAllSiteConfigs(confDir string) error {
	// 清理旧的站点专属配置文件
	// 每个站点生成自己的 geo/map 变量
	// 站点条件判断在 server 块中内联生成

	// 对于关联了规则的站点，生成站点专属的 geo/map 配置
	// （如果该站点有自己独立的 IP 黑名单或 Geo 规则，与全局规则不同的话）

	// 在新设计下，每个站点的规则通过 site_access_rules 关联
	// 站点专属变量使用 $site_{id}_blocked_ip 和 $site_{id}_geo_action
	// 但为了简化，如果站点关联了规则，直接在 server 块中使用全局变量
	// 因为全局配置已经包含了所有启用的规则

	// 清理旧的站点配置文件
	entries, err := os.ReadDir(confDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "site_") {
			os.Remove(filepath.Join(confDir, entry.Name()))
		}
	}

	return nil
}

// generateSiteConditions 生成站点访问控制条件判断
func (s *AccessRuleService) generateSiteConditions(siteID uint, merged *MergedAccessRules) string {
	var sb strings.Builder
	sb.WriteString("    # 访问控制规则\n")

	hasIP := len(merged.IPList) > 0
	hasGeo := len(merged.GeoRules) > 0

	if hasIP || hasGeo {
		// 生成站点专属的 geo/map 配置文件
		s.generateSiteSpecificConfig(siteID, merged)
	}

	if hasIP {
		sb.WriteString(fmt.Sprintf("    if ($site_%d_blocked_ip) { return 403; }\n", siteID))
	}
	if hasGeo {
		// 检查是否有 block 规则
		for _, action := range merged.GeoRules {
			if action == "block" {
				sb.WriteString(fmt.Sprintf("    if ($site_%d_geo_action = block) { return 403; }\n", siteID))
				break
			}
		}
	}

	return sb.String()
}

// generateSiteSpecificConfig 生成站点专属的 geo/map 配置文件
func (s *AccessRuleService) generateSiteSpecificConfig(siteID uint, merged *MergedAccessRules) {
	confDir := filepath.Join(config.AppConfig.Nginx.ConfDir, "access-control")

	// 生成站点 IP 黑名单 geo 配置
	if len(merged.IPList) > 0 {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("# 站点 %d IP 黑名单配置\n", siteID))
		sb.WriteString(fmt.Sprintf("geo $site_%d_blocked_ip {\n", siteID))
		sb.WriteString("    default 0;\n")
		for _, ip := range merged.IPList {
			sb.WriteString(fmt.Sprintf("    %s 1;\n", ip))
		}
		sb.WriteString("}\n")
		confPath := filepath.Join(confDir, fmt.Sprintf("site_%d_ip_blacklist.conf", siteID))
		if err := os.WriteFile(confPath, []byte(sb.String()), 0644); err != nil {
			log.Printf("写入站点 IP 黑名单配置失败: %v", err)
		}
	} else {
		// 生成空的 geo 配置
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("# 站点 %d IP 黑名单配置（空）\n", siteID))
		sb.WriteString(fmt.Sprintf("geo $site_%d_blocked_ip {\n", siteID))
		sb.WriteString("    default 0;\n")
		sb.WriteString("}\n")
		confPath := filepath.Join(confDir, fmt.Sprintf("site_%d_ip_blacklist.conf", siteID))
		os.WriteFile(confPath, []byte(sb.String()), 0644)
	}

	// 生成站点 Geo 规则 map 配置
	if len(merged.GeoRules) > 0 {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("# 站点 %d Geo 规则配置\n", siteID))
		sb.WriteString(fmt.Sprintf("map $geoip_country_code $site_%d_geo_action {\n", siteID))
		sb.WriteString("    default allow;\n")
		for cc, action := range merged.GeoRules {
			sb.WriteString(fmt.Sprintf("    %s %s;\n", cc, action))
		}
		sb.WriteString("}\n")
		confPath := filepath.Join(confDir, fmt.Sprintf("site_%d_geo.conf", siteID))
		if err := os.WriteFile(confPath, []byte(sb.String()), 0644); err != nil {
			log.Printf("写入站点 Geo 规则配置失败: %v", err)
		}
	} else {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("# 站点 %d Geo 规则配置（空）\n", siteID))
		sb.WriteString(fmt.Sprintf("map $remote_addr $site_%d_geo_action {\n", siteID))
		sb.WriteString("    default allow;\n")
		sb.WriteString("}\n")
		confPath := filepath.Join(confDir, fmt.Sprintf("site_%d_geo.conf", siteID))
		os.WriteFile(confPath, []byte(sb.String()), 0644)
	}
}

func (s *AccessRuleService) reloadNginx() {
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

// ==================== 辅助函数 ====================

func countRuleItems(items []model.AccessRuleItem) (ipCount, geoCount int) {
	for _, item := range items {
		switch item.ItemType {
		case "ip":
			ipCount++
		case "geo":
			geoCount++
		}
	}
	return
}

func validateItemDto(dto *AccessRuleItemDto) error {
	switch dto.ItemType {
	case "ip":
		if dto.IPAddress == "" {
			return fmt.Errorf("IP 条目缺少 IP 地址")
		}
		if !isValidIPorCIDR(dto.IPAddress) {
			return fmt.Errorf("无效的 IP 地址或 CIDR 格式: %s", dto.IPAddress)
		}
	case "geo":
		if dto.CountryCode == "" {
			return fmt.Errorf("Geo 条目缺少国家代码")
		}
		if len(dto.CountryCode) < 2 || len(dto.CountryCode) > 3 {
			return fmt.Errorf("无效的国家代码: %s", dto.CountryCode)
		}
		if dto.Action != "allow" && dto.Action != "block" {
			dto.Action = "block"
		}
	default:
		return fmt.Errorf("无效的条目类型: %s", dto.ItemType)
	}
	return nil
}

func dtoToItem(dto AccessRuleItemDto) model.AccessRuleItem {
	return model.AccessRuleItem{
		RuleID:      dto.RuleID,
		ItemType:    dto.ItemType,
		IPAddress:   dto.IPAddress,
		CountryCode: strings.ToUpper(dto.CountryCode),
		Action:      dto.Action,
		Note:        dto.Note,
	}
}

func ruleToDto(rule *model.AccessRule) *AccessRuleDto {
	dto := &AccessRuleDto{
		ID:          rule.ID,
		Name:        rule.Name,
		Description: rule.Description,
		Enabled:     rule.Enabled,
		Items:       make([]AccessRuleItemDto, len(rule.Items)),
	}
	for i, item := range rule.Items {
		dto.Items[i] = AccessRuleItemDto{
			ID:          item.ID,
			RuleID:      item.RuleID,
			ItemType:    item.ItemType,
			IPAddress:   item.IPAddress,
			CountryCode: item.CountryCode,
			Action:      item.Action,
			Note:        item.Note,
		}
	}
	return dto
}
