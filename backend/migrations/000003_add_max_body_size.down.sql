-- 回滚：移除最大上传大小字段
ALTER TABLE sites DROP COLUMN IF EXISTS max_body_size;
