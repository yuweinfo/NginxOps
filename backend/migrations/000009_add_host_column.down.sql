DROP INDEX IF EXISTS idx_access_log_host;
ALTER TABLE access_log DROP COLUMN IF EXISTS host;
