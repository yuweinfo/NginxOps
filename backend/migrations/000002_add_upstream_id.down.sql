-- 回滚：删除 sites 表的 upstream_id 字段
ALTER TABLE sites DROP COLUMN IF EXISTS upstream_id;
