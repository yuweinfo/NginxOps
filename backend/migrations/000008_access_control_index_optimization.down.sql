-- =====================================================
-- Access Control Performance Optimization - 回滚索引优化
-- =====================================================

-- =========================================
-- 1. 删除 access_rule_items 表索引
-- =========================================

DROP INDEX IF EXISTS idx_access_rule_items_rule_type;

-- =========================================
-- 2. 删除 site_access_rules 表索引
-- =========================================

DROP INDEX IF EXISTS idx_site_access_rules_site_rule;
