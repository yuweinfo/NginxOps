-- =====================================================
-- NginxOps v1.0.0 - Initial Schema
-- golang-migrate migration
-- =====================================================

-- =========================================
-- 1. Access Log (访问日志)
-- =========================================
CREATE TABLE access_log (
    id          BIGSERIAL PRIMARY KEY,
    remote_addr VARCHAR(45)     NOT NULL,
    remote_user VARCHAR(255),
    time_local  TIMESTAMPTZ     NOT NULL,
    request     VARCHAR(2048)   NOT NULL,
    method      VARCHAR(10),
    path        VARCHAR(1024),
    protocol    VARCHAR(10),
    status      INTEGER         NOT NULL DEFAULT 200,
    body_bytes  BIGINT          NOT NULL DEFAULT 0,
    referer     VARCHAR(2048),
    user_agent  VARCHAR(1024),
    rt          DOUBLE PRECISION DEFAULT 0,
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_access_log_time ON access_log (time_local DESC);
CREATE INDEX idx_access_log_status ON access_log (status);
CREATE INDEX idx_access_log_path ON access_log (path);
CREATE INDEX idx_access_log_ip ON access_log (remote_addr);
CREATE INDEX idx_access_log_time_status ON access_log (time_local, status);

-- =========================================
-- 2. IP Geo Cache (IP地理位置缓存)
-- =========================================
CREATE TABLE ip_geo_cache (
    ip          VARCHAR(45)     PRIMARY KEY,
    country     VARCHAR(100),
    region      VARCHAR(100),
    city        VARCHAR(100),
    isp         VARCHAR(255),
    lat         DOUBLE PRECISION,
    lon         DOUBLE PRECISION,
    updated_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- =========================================
-- 3. Nginx Config History (Nginx配置历史)
-- =========================================
CREATE TABLE nginx_config_history (
    id          BIGSERIAL PRIMARY KEY,
    config_name VARCHAR(255)    NOT NULL DEFAULT 'main',
    config_type VARCHAR(20)     NOT NULL DEFAULT 'nginx',
    file_path   VARCHAR(512)    NOT NULL,
    content     TEXT            NOT NULL,
    remark      VARCHAR(500),
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_nginx_config_history_created ON nginx_config_history (created_at DESC);

-- =========================================
-- 4. Stats Summary (统计汇总)
-- =========================================
CREATE TABLE stats_summary (
    id          BIGSERIAL PRIMARY KEY,
    stat_date   DATE            NOT NULL,
    pv          BIGINT          NOT NULL DEFAULT 0,
    uv          BIGINT          NOT NULL DEFAULT 0,
    status_2xx  BIGINT          NOT NULL DEFAULT 0,
    status_3xx  BIGINT          NOT NULL DEFAULT 0,
    status_4xx  BIGINT          NOT NULL DEFAULT 0,
    status_5xx  BIGINT          NOT NULL DEFAULT 0,
    avg_rt      DOUBLE PRECISION DEFAULT 0,
    max_rt      DOUBLE PRECISION DEFAULT 0,
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_stats_summary_date ON stats_summary (stat_date DESC);

-- =========================================
-- 5. Users (用户表)
-- =========================================
CREATE TABLE users (
    id          BIGSERIAL PRIMARY KEY,
    username    VARCHAR(50)     NOT NULL UNIQUE,
    password    VARCHAR(255)    NOT NULL,
    email       VARCHAR(100),
    role        VARCHAR(20)     NOT NULL DEFAULT 'user',
    enabled     BOOLEAN         NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- =========================================
-- 6. Sites (站点表)
-- =========================================
CREATE TABLE sites (
    id              BIGSERIAL PRIMARY KEY,
    file_name       VARCHAR(255),
    domain          VARCHAR(255)    NOT NULL,
    port            INTEGER         NOT NULL DEFAULT 80,
    site_type       VARCHAR(20)     NOT NULL DEFAULT 'proxy',
    root_dir        VARCHAR(512),
    locations       TEXT,
    upstream_servers TEXT,
    ssl_enabled     BOOLEAN         DEFAULT FALSE,
    cert_id         BIGINT,
    force_https     BOOLEAN         DEFAULT FALSE,
    gzip            BOOLEAN         DEFAULT FALSE,
    cache           BOOLEAN         DEFAULT FALSE,
    enabled         BOOLEAN         NOT NULL DEFAULT TRUE,
    config          TEXT,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sites_domain ON sites (domain);
CREATE INDEX idx_sites_enabled ON sites (enabled);

-- =========================================
-- 7. Certificates (证书表)
-- =========================================
CREATE TABLE certificates (
    id          BIGSERIAL PRIMARY KEY,
    domain      VARCHAR(255)    NOT NULL,
    issuer      VARCHAR(50),
    issued_at   TIMESTAMPTZ,
    expires_at  TIMESTAMPTZ,
    status      VARCHAR(20)     NOT NULL DEFAULT 'pending',
    auto_renew  BOOLEAN         NOT NULL DEFAULT TRUE,
    cert_path   VARCHAR(512),
    key_path    VARCHAR(512),
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_certificates_domain ON certificates (domain);
CREATE INDEX idx_certificates_status ON certificates (status);

-- =========================================
-- 8. Upstreams (负载均衡表)
-- =========================================
CREATE TABLE upstreams (
    id              BIGSERIAL PRIMARY KEY,
    name            VARCHAR(100)    NOT NULL UNIQUE,
    lb_mode         VARCHAR(20)     NOT NULL DEFAULT 'round_robin',
    health_check    BOOLEAN         NOT NULL DEFAULT FALSE,
    check_interval  INTEGER         DEFAULT 5,
    check_path      VARCHAR(255)    DEFAULT '/',
    check_timeout   INTEGER         DEFAULT 3,
    servers         TEXT            NOT NULL DEFAULT '[]',
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- =========================================
-- 9. DNS Providers (DNS服务商配置)
-- =========================================
CREATE TABLE dns_providers (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    provider_type VARCHAR(50) NOT NULL,
    access_key_id VARCHAR(255) NOT NULL,
    access_key_secret VARCHAR(255) NOT NULL,
    is_default BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(name)
);

COMMENT ON TABLE dns_providers IS 'DNS 服务商配置';
COMMENT ON COLUMN dns_providers.name IS '配置名称';
COMMENT ON COLUMN dns_providers.provider_type IS '服务商类型: aliyun, tencent, cloudflare';
COMMENT ON COLUMN dns_providers.access_key_id IS '访问密钥 ID';
COMMENT ON COLUMN dns_providers.access_key_secret IS '访问密钥 Secret';
COMMENT ON COLUMN dns_providers.is_default IS '是否为默认配置';

-- =========================================
-- 10. Certificate Requests (证书申请记录)
-- =========================================
CREATE TABLE certificate_requests (
    id BIGSERIAL PRIMARY KEY,
    certificate_id BIGINT NOT NULL REFERENCES certificates(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL,
    challenge_type VARCHAR(50) DEFAULT 'dns-01',
    dns_provider_id BIGINT REFERENCES dns_providers(id),
    dns_record_name VARCHAR(255),
    dns_record_value TEXT,
    error_message TEXT,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE certificate_requests IS '证书申请记录';
COMMENT ON COLUMN certificate_requests.certificate_id IS '关联的证书记录';
COMMENT ON COLUMN certificate_requests.status IS '申请状态: pending, validating, issuing, completed, failed';
COMMENT ON COLUMN certificate_requests.challenge_type IS '验证方式: dns-01';
COMMENT ON COLUMN certificate_requests.dns_provider_id IS '使用的 DNS 服务商配置';
COMMENT ON COLUMN certificate_requests.dns_record_name IS 'DNS TXT 记录名称';
COMMENT ON COLUMN certificate_requests.dns_record_value IS 'DNS TXT 记录值';
COMMENT ON COLUMN certificate_requests.error_message IS '错误信息';

CREATE INDEX idx_certificate_requests_cert_id ON certificate_requests(certificate_id);
CREATE INDEX idx_certificate_requests_status ON certificate_requests(status);

-- =========================================
-- 11. Audit Log (审计日志)
-- =========================================
CREATE TABLE audit_log (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT,
    username VARCHAR(50),
    action VARCHAR(50) NOT NULL,
    module VARCHAR(50) NOT NULL,
    target_type VARCHAR(50),
    target_id BIGINT,
    target_name VARCHAR(100),
    detail TEXT,
    ip_address VARCHAR(50),
    user_agent VARCHAR(255),
    status VARCHAR(20) NOT NULL DEFAULT 'SUCCESS',
    error_msg TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE audit_log IS '审计日志';
COMMENT ON COLUMN audit_log.user_id IS '用户ID';
COMMENT ON COLUMN audit_log.username IS '用户名';
COMMENT ON COLUMN audit_log.action IS '操作类型: LOGIN, CREATE, UPDATE, DELETE等';
COMMENT ON COLUMN audit_log.module IS '模块: AUTH, SITE, CERTIFICATE等';
COMMENT ON COLUMN audit_log.target_type IS '目标类型';
COMMENT ON COLUMN audit_log.target_id IS '目标ID';
COMMENT ON COLUMN audit_log.target_name IS '目标名称';
COMMENT ON COLUMN audit_log.detail IS '详情(JSON)';
COMMENT ON COLUMN audit_log.ip_address IS 'IP地址';
COMMENT ON COLUMN audit_log.user_agent IS 'User-Agent';
COMMENT ON COLUMN audit_log.status IS '状态: SUCCESS, FAILURE';
COMMENT ON COLUMN audit_log.error_msg IS '错误信息';

CREATE INDEX idx_audit_log_user_id ON audit_log(user_id);
CREATE INDEX idx_audit_log_module ON audit_log(module);
CREATE INDEX idx_audit_log_action ON audit_log(action);
CREATE INDEX idx_audit_log_created ON audit_log(created_at DESC);
CREATE INDEX idx_audit_log_status ON audit_log(status);
