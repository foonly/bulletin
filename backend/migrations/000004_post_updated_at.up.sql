-- 000004_post_updated_at.up.sql

ALTER TABLE posts ADD COLUMN updated_at TIMESTAMP WITH TIME ZONE;
