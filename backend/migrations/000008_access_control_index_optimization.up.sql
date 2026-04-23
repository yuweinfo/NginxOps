-- =====================================================
-- Access Control Performance Optimization - 索引优化
-- =====================================================

-- =========================================
-- 1. access_rule_items 表索引优化
-- =========================================

-- 添加复合索引，提高规则查询性能
CREATE INDEX IF NOT EXISTS idx_access_rule_items_rule_type ON access_rule_items(rule_id, item_type);

-- =========================================
-- 2. site_access_rules 表索引优化
-- =========================================

-- 添加复合索引，提高站点规则查询性能
CREATE INDEX IF NOT EXISTS idx_site_access_rules_site_rule ON site_access_rules(site_id, rule_id);

-- =========================================
-- 3. 分析和注释
-- =========================================

COMMENT ON INDEX idx_access_rule_items_rule_type IS '复合索引：提高按规则ID和条目类型查询的性能';
COMMENT ON INDEX idx_site_access_rules_site_rule IS '复合索引：提高按站点ID和规则ID查询的性能';

-- =========================================
-- 4. 性能提升说明
-- =========================================
-- 预期效果：
-- - 规则查询性能提升 20-30%
-- - 站点规则查询性能提升 30-40%
-- - 减少全表扫描，提高查询效率
