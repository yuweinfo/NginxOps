package repository

import (
	"nginxops/internal/database"
	"nginxops/internal/model"

	"gorm.io/gorm"
)

// AccessControlRepository 访问控制数据仓库
type AccessControlRepository struct{}

func NewAccessControlRepository() *AccessControlRepository {
	return &AccessControlRepository{}
}

// ==================== 全局设置 ====================

func (r *AccessControlRepository) GetSettings() (*model.AccessControlSettings, error) {
	var settings model.AccessControlSettings
	err := database.DB.First(&settings).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 创建默认设置
			settings = model.AccessControlSettings{
				GeoEnabled:         false,
				IPBlacklistEnabled: false,
				DefaultAction:      "allow",
			}
			if err := database.DB.Create(&settings).Error; err != nil {
				return nil, err
			}
			return &settings, nil
		}
		return nil, err
	}
	return &settings, nil
}

func (r *AccessControlRepository) UpdateSettings(settings *model.AccessControlSettings) error {
	return database.DB.Save(settings).Error
}

// ==================== 全局 IP 黑名单 ====================

func (r *AccessControlRepository) ListIPBlacklist() ([]model.IPBlacklist, error) {
	var list []model.IPBlacklist
	err := database.DB.Order("created_at DESC").Find(&list).Error
	return list, err
}

func (r *AccessControlRepository) FindIPBlacklistByID(id uint) (*model.IPBlacklist, error) {
	var item model.IPBlacklist
	err := database.DB.First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *AccessControlRepository) CreateIPBlacklist(item *model.IPBlacklist) error {
	return database.DB.Create(item).Error
}

func (r *AccessControlRepository) UpdateIPBlacklist(item *model.IPBlacklist) error {
	return database.DB.Save(item).Error
}

func (r *AccessControlRepository) DeleteIPBlacklist(id uint) error {
	return database.DB.Delete(&model.IPBlacklist{}, id).Error
}

func (r *AccessControlRepository) FindEnabledIPBlacklist() ([]model.IPBlacklist, error) {
	var list []model.IPBlacklist
	err := database.DB.Where("enabled = ?", true).Find(&list).Error
	return list, err
}

// ==================== 全局 Geo 规则 ====================

func (r *AccessControlRepository) ListGeoRules() ([]model.GeoRule, error) {
	var list []model.GeoRule
	err := database.DB.Order("country_code ASC").Find(&list).Error
	return list, err
}

func (r *AccessControlRepository) FindGeoRuleByID(id uint) (*model.GeoRule, error) {
	var item model.GeoRule
	err := database.DB.First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *AccessControlRepository) CreateGeoRule(item *model.GeoRule) error {
	return database.DB.Create(item).Error
}

func (r *AccessControlRepository) UpdateGeoRule(item *model.GeoRule) error {
	return database.DB.Save(item).Error
}

func (r *AccessControlRepository) DeleteGeoRule(id uint) error {
	return database.DB.Delete(&model.GeoRule{}, id).Error
}

func (r *AccessControlRepository) FindEnabledGeoRules() ([]model.GeoRule, error) {
	var list []model.GeoRule
	err := database.DB.Where("enabled = ?", true).Find(&list).Error
	return list, err
}

// ==================== 站点专属 IP 黑名单 ====================

func (r *AccessControlRepository) ListSiteIPBlacklist(siteID uint) ([]model.SiteIPBlacklist, error) {
	var list []model.SiteIPBlacklist
	err := database.DB.Where("site_id = ?", siteID).Order("created_at DESC").Find(&list).Error
	return list, err
}

func (r *AccessControlRepository) CreateSiteIPBlacklist(item *model.SiteIPBlacklist) error {
	return database.DB.Create(item).Error
}

func (r *AccessControlRepository) DeleteSiteIPBlacklist(id uint) error {
	return database.DB.Delete(&model.SiteIPBlacklist{}, id).Error
}

func (r *AccessControlRepository) FindEnabledSiteIPBlacklist(siteID uint) ([]model.SiteIPBlacklist, error) {
	var list []model.SiteIPBlacklist
	err := database.DB.Where("site_id = ? AND enabled = ?", siteID, true).Find(&list).Error
	return list, err
}

// ==================== 站点专属 Geo 规则 ====================

func (r *AccessControlRepository) ListSiteGeoRules(siteID uint) ([]model.SiteGeoRule, error) {
	var list []model.SiteGeoRule
	err := database.DB.Where("site_id = ?", siteID).Order("country_code ASC").Find(&list).Error
	return list, err
}

func (r *AccessControlRepository) CreateSiteGeoRule(item *model.SiteGeoRule) error {
	return database.DB.Create(item).Error
}

func (r *AccessControlRepository) DeleteSiteGeoRule(id uint) error {
	return database.DB.Delete(&model.SiteGeoRule{}, id).Error
}

func (r *AccessControlRepository) FindEnabledSiteGeoRules(siteID uint) ([]model.SiteGeoRule, error) {
	var list []model.SiteGeoRule
	err := database.DB.Where("site_id = ? AND enabled = ?", siteID, true).Find(&list).Error
	return list, err
}
