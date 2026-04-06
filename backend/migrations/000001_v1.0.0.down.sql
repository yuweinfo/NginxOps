-- =====================================================
-- NginxOps v1.0.0 - Rollback
-- golang-migrate migration - DOWN
-- =====================================================

-- 按照依赖关系逆序删除表
DROP TABLE IF EXISTS audit_log;
DROP TABLE IF EXISTS certificate_requests;
DROP TABLE IF EXISTS dns_providers;
DROP TABLE IF EXISTS upstreams;
DROP TABLE IF EXISTS certificates;
DROP TABLE IF EXISTS sites;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS stats_summary;
DROP TABLE IF EXISTS nginx_config_history;
DROP TABLE IF EXISTS ip_geo_cache;
DROP TABLE IF EXISTS access_log;
