-- =====================================================
-- Rollback: Access Control - Geo/IP 封锁功能
-- =====================================================

-- 删除站点专属规则表
DROP TABLE IF EXISTS site_geo_rules;
DROP TABLE IF EXISTS site_ip_blacklist;

-- 删除 sites 表中的访问控制模式字段
ALTER TABLE sites DROP COLUMN IF EXISTS access_control_mode;

-- 删除全局配置和规则表
DROP TABLE IF EXISTS access_control_settings;
DROP TABLE IF EXISTS geo_rules;
DROP TABLE IF EXISTS ip_blacklist;
