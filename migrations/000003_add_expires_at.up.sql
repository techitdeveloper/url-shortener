ALTER TABLE urls ADD COLUMN expires_at TIMESTAMP;

CREATE INDEX IF NOT EXISTS idx_urls_expires_at ON urls(expires_at)
WHERE expires_at IS NOT NULL;