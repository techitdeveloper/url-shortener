CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Index for fast email lookups
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Add user_id to urls table
ALTER TABLE urls ADD COLUMN user_id INTEGER;

-- Add foreign key constraint
ALTER TABLE urls ADD CONSTRAINT fk_urls_user_id 
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Index for finding user's URLs
CREATE INDEX IF NOT EXISTS idx_urls_user_id ON urls(user_id);