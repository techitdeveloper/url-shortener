-- Remove foreign key constraint first
ALTER TABLE urls DROP CONSTRAINT IF EXISTS fk_urls_user_id;

-- Remove user_id column from urls
ALTER TABLE urls DROP COLUMN IF EXISTS user_id;

-- Drop users table
DROP TABLE IF EXISTS users;