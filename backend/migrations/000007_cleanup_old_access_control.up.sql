-- 清理旧版访问控制系统的表和字段
-- 旧版功能已完全迁移到新版 access_rules 体系

-- 删除旧版访问控制相关表
DROP TABLE IF EXISTS site_geo_rules;
DROP TABLE IF EXISTS site_ip_blacklist;
DROP TABLE IF EXISTS access_control_settings;
DROP TABLE IF EXISTS geo_rules;
DROP TABLE IF EXISTS ip_blacklist;

-- 删除 sites 表中的旧版访问控制模式字段
ALTER TABLE sites DROP COLUMN IF EXISTS access_control_mode;
