CREATE TABLE IF NOT EXISTS follows (
    follower_id VARCHAR(256) NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    following_id VARCHAR(256) NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (follower_id, following_id),
    CONSTRAINT check_self_follow CHECK (follower_id <> following_id)
);

CREATE INDEX idx_follows_follower ON follows(follower_id);
CREATE INDEX idx_follows_following ON follows(following_id);