-- 000003_pinned_tags.up.sql

ALTER TABLE tags ADD COLUMN is_pinned BOOLEAN DEFAULT FALSE;
