-- 为 sites 表添加 upstream_id 字段，支持站点关联已定义的负载均衡器
ALTER TABLE sites ADD COLUMN upstream_id BIGINT REFERENCES upstreams(id) ON DELETE SET NULL;

COMMENT ON COLUMN sites.upstream_id IS '关联的负载均衡器ID';
