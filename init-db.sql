-- Table: urls (Stores URL Mappings)
CREATE TABLE IF NOT EXISTS urls (
    short_url VARCHAR(10) PRIMARY KEY,
    long_url TEXT UNIQUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ
);

-- Table: url_clicks (Tracks Click Events)
CREATE TABLE IF NOT EXISTS url_clicks (
    id BIGSERIAL PRIMARY KEY,
    short_url VARCHAR(10) NOT NULL REFERENCES urls(short_url) ON DELETE CASCADE,
    accessed_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for Performance Optimization
CREATE INDEX idx_url_clicks_time ON url_clicks(accessed_at);
CREATE INDEX idx_url_short_url ON url_clicks(short_url);
