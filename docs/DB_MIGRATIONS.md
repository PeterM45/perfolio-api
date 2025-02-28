-- Migration files for Perfolio API database schema

-- 000001_create_users_table.up.sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS users (
id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
email VARCHAR(255) NOT NULL UNIQUE,
display_name VARCHAR(100) NOT NULL,
bio TEXT,
avatar_url TEXT,
auth_id VARCHAR(255) UNIQUE,
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_auth_id ON users(auth_id);

COMMENT ON TABLE users IS 'User profiles for the Perfolio platform';

-- 000001_create_users_table.down.sql
DROP TABLE IF EXISTS users;

-- 000002_create_connections_table.up.sql
CREATE TABLE IF NOT EXISTS connections (
id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
follower_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
following_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
UNIQUE(follower_id, following_id)
);

CREATE INDEX idx_connections_follower ON connections(follower_id);
CREATE INDEX idx_connections_following ON connections(following_id);
CREATE INDEX idx_connections_pair ON connections(follower_id, following_id);

COMMENT ON TABLE connections IS 'Connections/follows between users';

-- 000002_create_connections_table.down.sql
DROP TABLE IF EXISTS connections;

-- 000003_create_posts_table.up.sql
CREATE TABLE IF NOT EXISTS posts (
id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
content TEXT NOT NULL,
url TEXT,
url_title TEXT,
url_description TEXT,
url_image TEXT,
published BOOLEAN NOT NULL DEFAULT TRUE,
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_posts_user_id ON posts(user_id);
CREATE INDEX idx_posts_created_at ON posts(created_at);
CREATE INDEX idx_posts_published_created_at ON posts(published, created_at DESC);

COMMENT ON TABLE posts IS 'User posts and shared content';

-- 000003_create_posts_table.down.sql
DROP TABLE IF EXISTS posts;

-- 000004_create_hashtags_table.up.sql
CREATE TABLE IF NOT EXISTS hashtags (
id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
name VARCHAR(50) NOT NULL UNIQUE,
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_hashtags_name ON hashtags(name);

CREATE TABLE IF NOT EXISTS post_hashtags (
post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
hashtag_id UUID NOT NULL REFERENCES hashtags(id) ON DELETE CASCADE,
PRIMARY KEY (post_id, hashtag_id)
);

CREATE INDEX idx_post_hashtags_post_id ON post_hashtags(post_id);
CREATE INDEX idx_post_hashtags_hashtag_id ON post_hashtags(hashtag_id);

COMMENT ON TABLE hashtags IS 'Hashtags for categorizing content';
COMMENT ON TABLE post_hashtags IS 'Junction table linking posts to hashtags';

-- 000004_create_hashtags_table.down.sql
DROP TABLE IF EXISTS post_hashtags;
DROP TABLE IF EXISTS hashtags;

-- 000005_create_reactions_table.up.sql
CREATE TYPE reaction_type AS ENUM ('like', 'celebrate', 'support', 'insightful', 'curious');

CREATE TABLE IF NOT EXISTS reactions (
id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
reaction_type reaction_type NOT NULL,
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
UNIQUE(user_id, post_id)
);

CREATE INDEX idx_reactions_post_id ON reactions(post_id);
CREATE INDEX idx_reactions_user_id ON reactions(user_id);

COMMENT ON TABLE reactions IS 'User reactions to posts';

-- 000005_create_reactions_table.down.sql
DROP TABLE IF EXISTS reactions;
DROP TYPE IF EXISTS reaction_type;

-- 000006_create_widgets_table.up.sql
CREATE TYPE widget_type AS ENUM (
'profile_info',
'recent_posts',
'connections',
'calendar',
'about',
'skills',
'experience',
'education',
'projects',
'custom'
);

CREATE TABLE IF NOT EXISTS widgets (
id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
widget_type widget_type NOT NULL,
title VARCHAR(100) NOT NULL,
content JSONB,
position JSONB NOT NULL,
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_widgets_user_id ON widgets(user_id);

COMMENT ON TABLE widgets IS 'Customizable profile widgets for users';
COMMENT ON COLUMN widgets.position IS 'JSON containing x, y, width, height properties for grid placement';
COMMENT ON COLUMN widgets.content IS 'Widget-specific content in JSON format';

-- 000006_create_widgets_table.down.sql
DROP TABLE IF EXISTS widgets;
DROP TYPE IF EXISTS widget_type;

-- 000007_add_widget_version.up.sql
-- Add version column for optimistic concurrency control
ALTER TABLE widgets ADD COLUMN version INTEGER NOT NULL DEFAULT 1;

-- 000007_add_widget_version.down.sql
ALTER TABLE widgets DROP COLUMN version;
