-- =====================================================
-- Access Control - Geo/IP 封锁功能
-- =====================================================

-- =========================================
-- 1. 全局 IP 黑名单
-- =========================================
CREATE TABLE ip_blacklist (
    id          BIGSERIAL PRIMARY KEY,
    ip_address  VARCHAR(100) NOT NULL,        -- IP 地址或 CIDR 网段
    note        VARCHAR(255),                  -- 备注说明
    enabled     BOOLEAN DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ip_blacklist_ip ON ip_blacklist(ip_address);
CREATE INDEX idx_ip_blacklist_enabled ON ip_blacklist(enabled);

COMMENT ON TABLE ip_blacklist IS '全局 IP 黑名单';
COMMENT ON COLUMN ip_blacklist.ip_address IS 'IP 地址或 CIDR 网段，如 192.168.1.1 或 10.0.0.0/8';
COMMENT ON COLUMN ip_blacklist.note IS '备注说明';

-- =========================================
-- 2. 全局 Geo 封锁规则
-- =========================================
CREATE TABLE geo_rules (
    id          BIGSERIAL PRIMARY KEY,
    country_code VARCHAR(10) NOT NULL,         -- 国家代码 (CN, US, RU 等)
    action      VARCHAR(10) NOT NULL DEFAULT 'block',  -- allow/block
    note        VARCHAR(255),
    enabled     BOOLEAN DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_geo_rules_country ON geo_rules(country_code);
CREATE INDEX idx_geo_rules_enabled ON geo_rules(enabled);
CREATE UNIQUE INDEX idx_geo_rules_country_unique ON geo_rules(country_code) WHERE enabled = TRUE;

COMMENT ON TABLE geo_rules IS '全局 Geo 封锁规则';
COMMENT ON COLUMN geo_rules.country_code IS '国家代码，如 CN, US, RU, KP 等';
COMMENT ON COLUMN geo_rules.action IS '动作: allow(允许) 或 block(封锁)';

-- =========================================
-- 3. 全局访问控制设置
-- =========================================
CREATE TABLE access_control_settings (
    id              BIGSERIAL PRIMARY KEY,
    geo_enabled     BOOLEAN DEFAULT FALSE,     -- 是否启用 Geo 封锁
    ip_blacklist_enabled BOOLEAN DEFAULT FALSE, -- 是否启用 IP 黑名单
    default_action  VARCHAR(10) DEFAULT 'allow', -- 默认动作 allow/block
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 插入默认配置
INSERT INTO access_control_settings (geo_enabled, ip_blacklist_enabled, default_action) 
VALUES (FALSE, FALSE, 'allow');

COMMENT ON TABLE access_control_settings IS '全局访问控制设置';
COMMENT ON COLUMN access_control_settings.geo_enabled IS '是否启用 Geo 封锁';
COMMENT ON COLUMN access_control_settings.ip_blacklist_enabled IS '是否启用 IP 黑名单';
COMMENT ON COLUMN access_control_settings.default_action IS '默认动作: allow(默认允许) 或 block(默认封锁)';

-- =========================================
-- 4. 站点访问控制模式（修改 sites 表）
-- =========================================
ALTER TABLE sites ADD COLUMN IF NOT EXISTS access_control_mode VARCHAR(20) DEFAULT 'inherit';
-- access_control_mode: 
--   'inherit'  - 继承全局规则（默认）
--   'merge'    - 全局 + 站点规则合并
--   'override' - 仅使用站点规则

COMMENT ON COLUMN sites.access_control_mode IS '访问控制模式: inherit(继承全局), merge(合并), override(覆盖)';

-- =========================================
-- 5. 站点专属 IP 黑名单
-- =========================================
CREATE TABLE site_ip_blacklist (
    id          BIGSERIAL PRIMARY KEY,
    site_id     BIGINT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    ip_address  VARCHAR(100) NOT NULL,
    note        VARCHAR(255),
    enabled     BOOLEAN DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_site_ip_blacklist_site ON site_ip_blacklist(site_id);
CREATE INDEX idx_site_ip_blacklist_enabled ON site_ip_blacklist(enabled);

COMMENT ON TABLE site_ip_blacklist IS '站点专属 IP 黑名单';

-- =========================================
-- 6. 站点专属 Geo 规则
-- =========================================
CREATE TABLE site_geo_rules (
    id          BIGSERIAL PRIMARY KEY,
    site_id     BIGINT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    country_code VARCHAR(10) NOT NULL,
    action      VARCHAR(10) NOT NULL DEFAULT 'block',
    note        VARCHAR(255),
    enabled     BOOLEAN DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_site_geo_rules_site ON site_geo_rules(site_id);
CREATE INDEX idx_site_geo_rules_enabled ON site_geo_rules(enabled);
CREATE UNIQUE INDEX idx_site_geo_rules_unique ON site_geo_rules(site_id, country_code) WHERE enabled = TRUE;

COMMENT ON TABLE site_geo_rules IS '站点专属 Geo 规则';
