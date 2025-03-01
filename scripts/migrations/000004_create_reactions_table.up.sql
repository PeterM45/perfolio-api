CREATE TYPE reaction_type AS ENUM ('like', 'celebrate', 'support', 'insightful', 'curious');

CREATE TABLE IF NOT EXISTS reaction (
    id VARCHAR(256) PRIMARY KEY,
    post_id VARCHAR(256) NOT NULL REFERENCES post(id) ON DELETE CASCADE,
    user_id VARCHAR(256) NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    type reaction_type NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(post_id, user_id)
);

CREATE INDEX idx_reaction_post_id ON reaction(post_id);
CREATE INDEX idx_reaction_user_id ON reaction(user_id);