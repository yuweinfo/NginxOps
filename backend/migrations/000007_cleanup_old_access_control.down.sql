-- 恢复旧版访问控制系统的表和字段（回滚用）

-- 恢复 sites 表中的 access_control_mode 字段
ALTER TABLE sites ADD COLUMN IF NOT EXISTS access_control_mode VARCHAR(20) DEFAULT 'custom';

-- 恢复旧版访问控制相关表（空表，数据不恢复）
CREATE TABLE IF NOT EXISTS ip_blacklist (
    id SERIAL PRIMARY KEY,
    ip_address VARCHAR(100) NOT NULL,
    note VARCHAR(255),
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS geo_rules (
    id SERIAL PRIMARY KEY,
    country_code VARCHAR(10) NOT NULL,
    action VARCHAR(10) NOT NULL DEFAULT 'block',
    note VARCHAR(255),
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS access_control_settings (
    id SERIAL PRIMARY KEY,
    geo_enabled BOOLEAN DEFAULT FALSE,
    ip_blacklist_enabled BOOLEAN DEFAULT FALSE,
    default_action VARCHAR(10) DEFAULT 'allow',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS site_ip_blacklist (
    id SERIAL PRIMARY KEY,
    site_id INTEGER NOT NULL,
    ip_address VARCHAR(100) NOT NULL,
    note VARCHAR(255),
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS site_geo_rules (
    id SERIAL PRIMARY KEY,
    site_id INTEGER NOT NULL,
    country_code VARCHAR(10) NOT NULL,
    action VARCHAR(10) NOT NULL DEFAULT 'block',
    note VARCHAR(255),
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
