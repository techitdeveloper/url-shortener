CREATE TABLE IF NOT EXISTS url_clicks (
    id SERIAL PRIMARY KEY,
    url_id INTEGER NOT NULL REFERENCES urls(id) ON DELETE CASCADE,
    clicked_at TIMESTAMP NOT NULL DEFAULT NOW(),
    ip_address VARCHAR(45),  -- IPv6 can be up to 45 chars
    user_agent TEXT,
    referer TEXT,
    country VARCHAR(2),      -- ISO country code (e.g., 'US', 'IN')
    city VARCHAR(100)
);

-- Index for finding clicks by URL
CREATE INDEX IF NOT EXISTS idx_url_clicks_url_id ON url_clicks(url_id);

-- Index for time-based queries
CREATE INDEX IF NOT EXISTS idx_url_clicks_clicked_at ON url_clicks(clicked_at);

-- Composite index for url + time range queries
CREATE INDEX IF NOT EXISTS idx_url_clicks_url_time ON url_clicks(url_id, clicked_at DESC);