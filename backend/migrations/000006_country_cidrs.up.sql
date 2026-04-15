-- =====================================================
-- Add Country CIDR Cache Table
-- =====================================================

CREATE TABLE country_cidrs (
    id           BIGSERIAL PRIMARY KEY,
    country_code VARCHAR(10) NOT NULL,
    cidrs        TEXT,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_country_cidrs_code ON country_cidrs(country_code);

COMMENT ON TABLE country_cidrs IS '国家IP段缓存，用于Geo访问控制';
COMMENT ON COLUMN country_cidrs.country_code IS '国家代码（如 CN, US）';
COMMENT ON COLUMN country_cidrs.cidrs IS 'JSON数组格式的CIDR列表';
