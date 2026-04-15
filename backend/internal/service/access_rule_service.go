package service

import (
	"fmt"
	"log"
	"net"
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
	Action      string `json:"action"`      // allow / block（IP 和 Geo 均使用）
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
	BlockedIPs []string          // 合并后的 IP 黑名单列表（action=block）
	AllowedIPs []string          // 合并后的 IP 白名单列表（action=allow）
	GeoRules   map[string]string // 合并后的 Geo 规则（countryCode -> action）
}

// MergeRules 合并多条规则为一个统一的规则集
// 逻辑：
// - IP规则：按 action 分类，block 和 allow 分别收集（去重）
//   - 如果同一 IP 同时出现在 allow 和 block 中，block 优先（安全优先）
// - Geo规则：相同国家代码，如果任一规则 block 则 block；如果所有规则都 allow 则 allow
//   优先级：block > allow（安全优先）
func (s *AccessRuleService) MergeRules(rules []model.AccessRule) *MergedAccessRules {
	merged := &MergedAccessRules{
		BlockedIPs: []string{},
		AllowedIPs: []string{},
		GeoRules:   make(map[string]string),
	}

	blockedSet := make(map[string]bool)
	allowedSet := make(map[string]bool)

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}
		for _, item := range rule.Items {
			switch item.ItemType {
			case "ip":
				if item.IPAddress != "" {
					action := item.Action
					if action == "" {
						action = "block" // 默认 block
					}
					if action == "block" {
						blockedSet[item.IPAddress] = true
						delete(allowedSet, item.IPAddress) // block 优先，从 allow 中移除
					} else if action == "allow" {
						if !blockedSet[item.IPAddress] { // 如果已被 block 则不加入 allow
							allowedSet[item.IPAddress] = true
						}
					}
				}
			case "geo":
				if item.CountryCode != "" {
					cc := strings.ToUpper(item.CountryCode)
					action := item.Action
					if action == "" {
						action = "block"
					}
					existing, exists := merged.GeoRules[cc]
					if !exists {
						merged.GeoRules[cc] = action
					} else {
						// 安全优先：block 覆盖 allow
						if existing == "allow" && action == "block" {
							merged.GeoRules[cc] = "block"
						}
					}
				}
			}
		}
	}

	for ip := range blockedSet {
		merged.BlockedIPs = append(merged.BlockedIPs, ip)
	}
	for ip := range allowedSet {
		merged.AllowedIPs = append(merged.AllowedIPs, ip)
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

	// 生成全局 IP 访问控制配置
	if err := s.generateIPAccessConfig(confDir, merged.BlockedIPs, merged.AllowedIPs); err != nil {
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

func (s *AccessRuleService) generateIPAccessConfig(confDir string, blockedIPs []string, allowedIPs []string) error {
	var sb strings.Builder
	sb.WriteString("# 自动生成的 IP 访问控制配置 - 请勿手动修改\n\n")

	// IP 黑名单
	if len(blockedIPs) > 0 {
		sb.WriteString("geo $global_blocked_ip {\n")
		sb.WriteString("    default 0;\n")
		for _, ip := range blockedIPs {
			sb.WriteString(fmt.Sprintf("    %s 1;\n", ip))
		}
		sb.WriteString("}\n")
	} else {
		sb.WriteString("geo $global_blocked_ip {\n")
		sb.WriteString("    default 0;\n")
		sb.WriteString("}\n")
	}

	// IP 白名单
	if len(allowedIPs) > 0 {
		sb.WriteString("\ngeo $global_allowed_ip {\n")
		sb.WriteString("    default 0;\n")
		for _, ip := range allowedIPs {
			sb.WriteString(fmt.Sprintf("    %s 1;\n", ip))
		}
		sb.WriteString("}\n")
	} else {
		sb.WriteString("\ngeo $global_allowed_ip {\n")
		sb.WriteString("    default 0;\n")
		sb.WriteString("}\n")
	}

	confPath := filepath.Join(confDir, "global_ip_access.conf")
	return os.WriteFile(confPath, []byte(sb.String()), 0644)
}

func (s *AccessRuleService) generateGeoConfig(confDir string, geoRules map[string]string) error {
	var sb strings.Builder
	sb.WriteString("# 自动生成的 Geo 封锁配置 - 请勿手动修改\n")
	sb.WriteString("# 使用 auth_request 委托后端进行 GeoIP 查询和访问控制判断\n\n")

	sb.WriteString("geo $global_geo_action {\n")
	sb.WriteString("    default allow;\n")
	sb.WriteString("}\n")

	confPath := filepath.Join(confDir, "global_geo.conf")
	return os.WriteFile(confPath, []byte(sb.String()), 0644)
}

func (s *AccessRuleService) syncAllSiteConfigs(confDir string) error {
	// 清理旧的站点配置文件
	entries, err := os.ReadDir(confDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		name := entry.Name()
		// 删除旧的配置文件
		if strings.HasPrefix(name, "site_") && (strings.HasSuffix(name, "_geo.conf") || strings.HasSuffix(name, "_ip_blacklist.conf") || strings.HasSuffix(name, "_ip_access.conf")) {
			os.Remove(filepath.Join(confDir, name))
		}
		// 删除旧的全局文件（已移到 nginx.conf 主配置中）
		if name == "global_ip_blacklist.conf" || name == "_global_real_ip_map.conf" {
			os.Remove(filepath.Join(confDir, name))
		}
	}

	return nil
}

// generateSiteConditions 生成站点访问控制条件判断
// 策略（network_mode: host 下 $remote_addr 即为真实客户端 IP）：
// - 仅 IP 黑名单：Nginx 层 geo+if 快速拦截（纯 Nginx，高性能）
// - 有 IP 白名单和/或 Geo 规则：通过 auth_request 后端统一检查
func (s *AccessRuleService) generateSiteConditions(siteID uint, merged *MergedAccessRules) string {
	var sb strings.Builder
	sb.WriteString("    # 访问控制规则\n")

	hasBlockedIPs := len(merged.BlockedIPs) > 0
	hasAllowedIPs := len(merged.AllowedIPs) > 0
	hasGeo := len(merged.GeoRules) > 0
	hasAnyRule := hasBlockedIPs || hasAllowedIPs || hasGeo

	if !hasAnyRule {
		return ""
	}

	// 生成站点专属 IP 配置文件（geo 指令必须在 http 块中）
	if hasBlockedIPs || hasAllowedIPs {
		s.generateSiteIPConfig(siteID, merged)
	}

	onlyBlockedIPs := hasBlockedIPs && !hasAllowedIPs && !hasGeo

	if onlyBlockedIPs {
		// 简单场景：仅 IP 黑名单，Nginx 层直接拦截（高性能）
		sb.WriteString("    # IP 黑名单 - 直接拒绝\n")
		sb.WriteString(fmt.Sprintf("    if ($site_%d_blocked_ip) { return 403; }\n", siteID))
	} else {
		// 复杂场景：有白名单和/或 Geo 规则，通过 auth_request 后端统一检查
		sb.WriteString("    # 访问控制 - 通过后端 API 统一检查\n")
		sb.WriteString("    # 后端处理优先级：IP 白名单 > IP 黑名单 > Geo 规则\n")
		sb.WriteString(fmt.Sprintf("    auth_request /access-control-check/site/%d;\n", siteID))

		// Nginx 层黑名单快速拦截（作为额外保障）
		if hasBlockedIPs {
			sb.WriteString("    # IP 黑名单 - Nginx 层快速拦截\n")
			sb.WriteString(fmt.Sprintf("    if ($site_%d_blocked_ip) { return 403; }\n", siteID))
		}

		// auth_request 代理端点
		sb.WriteString("\n    # auth_request 代理端点（internal）\n")
		sb.WriteString("    location /access-control-check/ {\n")
		sb.WriteString("        internal;\n")
		sb.WriteString("        auth_request off;\n")
		sb.WriteString("        proxy_pass http://127.0.0.1:8080/api/access-control/check/;\n")
		sb.WriteString("        proxy_set_header Host $host;\n")
		sb.WriteString("        proxy_set_header X-Real-IP $remote_addr;\n")
		sb.WriteString("        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;\n")
		sb.WriteString("        proxy_set_header X-Original-URI $request_uri;\n")
		sb.WriteString("        proxy_connect_timeout 5s;\n")
		sb.WriteString("        proxy_read_timeout 5s;\n")
		sb.WriteString("    }\n")
	}

	return sb.String()
}

// generateSiteIPConfig 生成站点 IP 访问控制 geo 配置文件
// network_mode: host 下 $remote_addr 即为真实客户端 IP，geo 直接基于 $remote_addr
func (s *AccessRuleService) generateSiteIPConfig(siteID uint, merged *MergedAccessRules) {
	confDir := filepath.Join(config.AppConfig.Nginx.ConfDir, "access-control")

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# 站点 %d IP 访问控制配置\n\n", siteID))

	// 黑名单 geo
	sb.WriteString(fmt.Sprintf("geo $site_%d_blocked_ip {\n", siteID))
	sb.WriteString("    default 0;\n")
	for _, ip := range merged.BlockedIPs {
		sb.WriteString(fmt.Sprintf("    %s 1;\n", ip))
	}
	sb.WriteString("}\n\n")

	// 白名单 geo
	sb.WriteString(fmt.Sprintf("geo $site_%d_allowed_ip {\n", siteID))
	sb.WriteString("    default 0;\n")
	for _, ip := range merged.AllowedIPs {
		sb.WriteString(fmt.Sprintf("    %s 1;\n", ip))
	}
	sb.WriteString("}\n")

	confPath := filepath.Join(confDir, fmt.Sprintf("site_%d_ip_access.conf", siteID))
	if err := os.WriteFile(confPath, []byte(sb.String()), 0644); err != nil {
		log.Printf("写入站点 IP 访问控制配置失败: %v", err)
	}

	// 清理旧的配置文件
	os.Remove(filepath.Join(confDir, fmt.Sprintf("site_%d_ip_blacklist.conf", siteID)))
	os.Remove(filepath.Join(confDir, fmt.Sprintf("site_%d_geo.conf", siteID)))
	os.Remove(filepath.Join(confDir, "_global_real_ip_map.conf"))
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

// ==================== 实时访问检查（auth_request 用） ====================

// CheckSiteAccessBlocked 检查指定 IP 是否被站点的访问控制规则阻止
// 逻辑：
// 1. 先检查 IP 白名单（allow），如果在白名单中则直接放行
// 2. 再检查 IP 黑名单（block），如果在黑名单中则拒绝
// 3. 最后检查 Geo 规则
// 返回: (是否阻止, 阻止原因, 错误)
func (s *AccessRuleService) CheckSiteAccessBlocked(clientIP string, siteIDStr string) (bool, string, error) {
	// 解析 siteID
	var siteID uint
	if _, err := fmt.Sscanf(siteIDStr, "%d", &siteID); err != nil {
		return false, "", err
	}

	// 获取站点的合并规则
	merged, err := s.GetSiteMergedRules(siteID)
	if err != nil {
		return false, "", err
	}

	// 1. 检查 IP 白名单（优先级最高）
	for _, allowedIP := range merged.AllowedIPs {
		if clientIP == allowedIP {
			return false, "", nil // 白名单放行
		}
		if strings.Contains(allowedIP, "/") {
			if ipInCIDR(clientIP, allowedIP) {
				return false, "", nil // 白名单放行
			}
		}
	}

	// 2. 检查 IP 黑名单
	for _, blockedIP := range merged.BlockedIPs {
		if clientIP == blockedIP {
			return true, "IP blacklist", nil
		}
		if strings.Contains(blockedIP, "/") {
			if ipInCIDR(clientIP, blockedIP) {
				return true, "IP blacklist (CIDR)", nil
			}
		}
	}

	// 3. 检查 Geo 规则
	if len(merged.GeoRules) > 0 {
		geoSvc := NewGeoIpService()
		geoInfo := geoSvc.GetGeo(clientIP)
		if geoInfo != nil && geoInfo.Country != "Unknown" {
			countryCode := geoSvc.GetCountryCode(geoInfo.Country)
			if action, exists := merged.GeoRules[strings.ToUpper(countryCode)]; exists {
				if action == "block" {
					return true, "Geo block: " + countryCode, nil
				}
			}
		}
	}

	return false, "", nil
}

// ipInCIDR 检查 IP 是否在 CIDR 范围内
func ipInCIDR(ipStr string, cidrStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	_, ipNet, err := net.ParseCIDR(cidrStr)
	if err != nil {
		return false
	}
	return ipNet.Contains(ip)
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
		// IP 条目也需要 action
		if dto.Action != "allow" && dto.Action != "block" {
			dto.Action = "block" // 默认 block
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

func isValidIPorCIDR(s string) bool {
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
