-- =====================================================
-- Rollback: Access Rule Refactor
-- =====================================================

-- 删除站点-规则关联表
DROP TABLE IF EXISTS site_access_rules;

-- 删除规则条目表
DROP TABLE IF EXISTS access_rule_items;

-- 删除规则主表
DROP TABLE IF EXISTS access_rules;

-- 恢复 sites 表的 access_control_mode 默认值
UPDATE sites SET access_control_mode = 'inherit' WHERE access_control_mode = 'custom';
