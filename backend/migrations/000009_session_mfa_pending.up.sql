-- 000009_session_mfa_pending.up.sql
ALTER TABLE sessions ADD COLUMN mfa_pending BOOLEAN NOT NULL DEFAULT FALSE;
