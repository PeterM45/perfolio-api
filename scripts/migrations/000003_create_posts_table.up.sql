CREATE TABLE IF NOT EXISTS post (
    id VARCHAR(256) PRIMARY KEY,
    user_id VARCHAR(256) NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    embed_urls TEXT[],
    hashtags TEXT[],
    visibility VARCHAR(32) NOT NULL DEFAULT 'public',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ
);

CREATE INDEX idx_post_user_id ON post(user_id);
CREATE INDEX idx_post_created_at ON post(created_at);
CREATE INDEX idx_post_visibility_created_at ON post(visibility, created_at DESC);