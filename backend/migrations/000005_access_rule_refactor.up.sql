-- =====================================================
-- Access Rule Refactor - 规则列表模式
-- =====================================================

-- =========================================
-- 1. 访问控制规则主表
-- =========================================
CREATE TABLE access_rules (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,            -- 规则名称
    description VARCHAR(255),                      -- 规则描述
    enabled     BOOLEAN DEFAULT TRUE,              -- 是否启用
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_access_rules_enabled ON access_rules(enabled);

COMMENT ON TABLE access_rules IS '访问控制规则 - 每条规则是独立的、可复用的单元';
COMMENT ON COLUMN access_rules.name IS '规则名称';
COMMENT ON COLUMN access_rules.description IS '规则描述';
COMMENT ON COLUMN access_rules.enabled IS '是否启用';

-- =========================================
-- 2. 规则条目表
-- =========================================
CREATE TABLE access_rule_items (
    id           BIGSERIAL PRIMARY KEY,
    rule_id      BIGINT NOT NULL REFERENCES access_rules(id) ON DELETE CASCADE,
    item_type    VARCHAR(20) NOT NULL,             -- ip / geo
    ip_address   VARCHAR(100),                      -- 当 item_type=ip 时使用
    country_code VARCHAR(10),                       -- 当 item_type=geo 时使用
    action       VARCHAR(10) DEFAULT 'block',       -- 当 item_type=geo 时使用: allow/block
    note         VARCHAR(255),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_access_rule_items_rule ON access_rule_items(rule_id);
CREATE INDEX idx_access_rule_items_type ON access_rule_items(item_type);

COMMENT ON TABLE access_rule_items IS '访问控制规则条目';
COMMENT ON COLUMN access_rule_items.item_type IS '条目类型: ip(IP黑名单) 或 geo(地理位置)';
COMMENT ON COLUMN access_rule_items.ip_address IS 'IP 地址或 CIDR 网段（item_type=ip 时使用）';
COMMENT ON COLUMN access_rule_items.country_code IS '国家代码（item_type=geo 时使用）';
COMMENT ON COLUMN access_rule_items.action IS '动作: allow(允许) 或 block(封锁)（item_type=geo 时使用）';

-- =========================================
-- 3. 站点-规则关联表
-- =========================================
CREATE TABLE site_access_rules (
    id         BIGSERIAL PRIMARY KEY,
    site_id    BIGINT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    rule_id    BIGINT NOT NULL REFERENCES access_rules(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_site_access_rules_unique ON site_access_rules(site_id, rule_id);
CREATE INDEX idx_site_access_rules_site ON site_access_rules(site_id);
CREATE INDEX idx_site_access_rules_rule ON site_access_rules(rule_id);

COMMENT ON TABLE site_access_rules IS '站点-规则关联表（多对多）';

-- =========================================
-- 4. 迁移旧数据：将旧的全局 IP 黑名单和 Geo 规则转换为规则列表
-- =========================================

-- 迁移 IP 黑名单为一条规则
INSERT INTO access_rules (name, description, enabled)
SELECT '全局IP黑名单', '从旧版全局IP黑名单迁移', TRUE
WHERE EXISTS (SELECT 1 FROM ip_blacklist LIMIT 1);

INSERT INTO access_rule_items (rule_id, item_type, ip_address, note, action)
SELECT 
    (SELECT id FROM access_rules WHERE name = '全局IP黑名单' LIMIT 1),
    'ip',
    ip_address,
    note,
    'block'
FROM ip_blacklist
WHERE EXISTS (SELECT 1 FROM access_rules WHERE name = '全局IP黑名单');

-- 迁移 Geo 规则为一条规则
INSERT INTO access_rules (name, description, enabled)
SELECT '全局Geo规则', '从旧版全局Geo规则迁移', TRUE
WHERE EXISTS (SELECT 1 FROM geo_rules LIMIT 1);

INSERT INTO access_rule_items (rule_id, item_type, ip_address, country_code, action, note)
SELECT 
    (SELECT id FROM access_rules WHERE name = '全局Geo规则' LIMIT 1),
    'geo',
    '',
    country_code,
    action,
    note
FROM geo_rules
WHERE EXISTS (SELECT 1 FROM access_rules WHERE name = '全局Geo规则');

-- 迁移站点专属规则为独立规则并关联到站点
-- 为每个有站点专属IP黑名单的站点创建规则
INSERT INTO access_rules (name, description, enabled)
SELECT 
    '站点' || site_id || '-IP黑名单',
    '从旧版站点专属IP黑名单迁移',
    TRUE
FROM site_ip_blacklist
GROUP BY site_id;

-- 为每个有站点专属Geo规则的站点创建规则
INSERT INTO access_rules (name, description, enabled)
SELECT 
    '站点' || site_id || '-Geo规则',
    '从旧版站点专属Geo规则迁移',
    TRUE
FROM site_geo_rules
GROUP BY site_id;

-- =========================================
-- 5. 更新 sites 表的 access_control_mode 默认值
-- =========================================
-- 新设计不再需要 inherit/merge/override 模式
-- 改为 custom 模式：站点通过关联规则来控制访问
UPDATE sites SET access_control_mode = 'custom' WHERE access_control_mode IN ('inherit', 'merge', 'override');
