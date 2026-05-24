-- 000002_read_markers.up.sql

CREATE TABLE read_markers (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL, -- Can be a circle_id (for chat) or post_id (for threads)
    last_read_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (user_id, entity_id)
);

-- Add index to speed up lookups
CREATE INDEX idx_read_markers_user_id ON read_markers(user_id);
