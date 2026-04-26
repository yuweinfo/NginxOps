ALTER TABLE access_log ADD COLUMN IF NOT EXISTS host VARCHAR(255) DEFAULT '';
CREATE INDEX IF NOT EXISTS idx_access_log_host ON access_log(host);
