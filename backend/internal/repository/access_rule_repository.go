package repository

import (
	"nginxops/internal/database"
	"nginxops/internal/model"

	"gorm.io/gorm"
)

// AccessRuleRepository 访问规则数据仓库
type AccessRuleRepository struct{}

func NewAccessRuleRepository() *AccessRuleRepository {
	return &AccessRuleRepository{}
}

// ==================== 规则 CRUD ====================

func (r *AccessRuleRepository) ListRules() ([]model.AccessRule, error) {
	var rules []model.AccessRule
	err := database.DB.Preload("Items", func(db *gorm.DB) *gorm.DB {
		return db.Order("id ASC")
	}).Order("created_at DESC").Find(&rules).Error
	return rules, err
}

func (r *AccessRuleRepository) FindRuleByID(id uint) (*model.AccessRule, error) {
	var rule model.AccessRule
	err := database.DB.Preload("Items", func(db *gorm.DB) *gorm.DB {
		return db.Order("id ASC")
	}).First(&rule, id).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *AccessRuleRepository) CreateRule(rule *model.AccessRule) error {
	return database.DB.Create(rule).Error
}

func (r *AccessRuleRepository) UpdateRule(rule *model.AccessRule) error {
	return database.DB.Save(rule).Error
}

func (r *AccessRuleRepository) DeleteRule(id uint) error {
	// 先删除关联的 items
	if err := database.DB.Where("rule_id = ?", id).Delete(&model.AccessRuleItem{}).Error; err != nil {
		return err
	}
	// 删除关联的 site_access_rules
	if err := database.DB.Where("rule_id = ?", id).Delete(&model.SiteAccessRule{}).Error; err != nil {
		return err
	}
	return database.DB.Delete(&model.AccessRule{}, id).Error
}

func (r *AccessRuleRepository) FindEnabledRules() ([]model.AccessRule, error) {
	var rules []model.AccessRule
	err := database.DB.Preload("Items", func(db *gorm.DB) *gorm.DB {
		return db.Order("id ASC")
	}).Where("enabled = ?", true).Find(&rules).Error
	return rules, err
}

func (r *AccessRuleRepository) FindEnabledRulesByIDs(ids []uint) ([]model.AccessRule, error) {
	var rules []model.AccessRule
	if len(ids) == 0 {
		return rules, nil
	}
	err := database.DB.Preload("Items", func(db *gorm.DB) *gorm.DB {
		return db.Order("id ASC")
	}).Where("id IN ? AND enabled = ?", ids, true).Find(&rules).Error
	return rules, err
}

// ==================== 规则条目 CRUD ====================

func (r *AccessRuleRepository) CreateItem(item *model.AccessRuleItem) error {
	return database.DB.Create(item).Error
}

func (r *AccessRuleRepository) UpdateItem(item *model.AccessRuleItem) error {
	return database.DB.Save(item).Error
}

func (r *AccessRuleRepository) DeleteItem(id uint) error {
	return database.DB.Delete(&model.AccessRuleItem{}, id).Error
}

func (r *AccessRuleRepository) DeleteItemsByRuleID(ruleID uint) error {
	return database.DB.Where("rule_id = ?", ruleID).Delete(&model.AccessRuleItem{}).Error
}

func (r *AccessRuleRepository) FindItemByID(id uint) (*model.AccessRuleItem, error) {
	var item model.AccessRuleItem
	err := database.DB.First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// ==================== 站点-规则关联 ====================

func (r *AccessRuleRepository) GetSiteRuleIDs(siteID uint) ([]uint, error) {
	var sarList []model.SiteAccessRule
	err := database.DB.Where("site_id = ?", siteID).Find(&sarList).Error
	if err != nil {
		return nil, err
	}
	ruleIDs := make([]uint, len(sarList))
	for i, sar := range sarList {
		ruleIDs[i] = sar.RuleID
	}
	return ruleIDs, nil
}

func (r *AccessRuleRepository) SetSiteRules(siteID uint, ruleIDs []uint) error {
	// 先删除旧的关联
	if err := database.DB.Where("site_id = ?", siteID).Delete(&model.SiteAccessRule{}).Error; err != nil {
		return err
	}
	// 创建新的关联
	for _, ruleID := range ruleIDs {
		sar := model.SiteAccessRule{SiteID: siteID, RuleID: ruleID}
		if err := database.DB.Create(&sar).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *AccessRuleRepository) GetRulesBySiteID(siteID uint) ([]model.AccessRule, error) {
	ruleIDs, err := r.GetSiteRuleIDs(siteID)
	if err != nil {
		return nil, err
	}
	if len(ruleIDs) == 0 {
		return []model.AccessRule{}, nil
	}
	return r.FindEnabledRulesByIDs(ruleIDs)
}

// ==================== 规则被引用情况 ====================

func (r *AccessRuleRepository) GetRuleSiteCount(ruleID uint) (int64, error) {
	var count int64
	err := database.DB.Model(&model.SiteAccessRule{}).Where("rule_id = ?", ruleID).Count(&count).Error
	return count, err
}

func (r *AccessRuleRepository) GetRulesSiteCounts() (map[uint]int64, error) {
	var results []struct {
		RuleID uint
		Count  int64
	}
	err := database.DB.Model(&model.SiteAccessRule{}).
		Select("rule_id, count(*) as count").
		Group("rule_id").
		Find(&results).Error
	if err != nil {
		return nil, err
	}
	counts := make(map[uint]int64)
	for _, r := range results {
		counts[r.RuleID] = r.Count
	}
	return counts, nil
}
