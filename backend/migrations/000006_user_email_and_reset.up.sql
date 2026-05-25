-- 000006_user_email_and_reset.up.sql
ALTER TABLE users ADD COLUMN email TEXT UNIQUE;

CREATE TABLE password_reset_tokens (
    token TEXT PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
