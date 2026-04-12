-- 添加最大上传大小字段
ALTER TABLE sites ADD COLUMN IF NOT EXISTS max_body_size INTEGER DEFAULT 200;

COMMENT ON COLUMN sites.max_body_size IS '最大上传大小（MB），默认200MB';
