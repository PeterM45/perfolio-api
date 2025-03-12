-- Add missing columns
ALTER TABLE widgets ADD COLUMN version INTEGER NOT NULL DEFAULT 1;
ALTER TABLE widgets ADD COLUMN is_visible BOOLEAN NOT NULL DEFAULT true;
ALTER TABLE widgets ADD COLUMN display_name VARCHAR(256);
ALTER TABLE widgets ADD COLUMN deleted_at TIMESTAMP WITH TIME ZONE;

-- Convert settings to JSONB for better querying
ALTER TABLE widgets ALTER COLUMN settings TYPE JSONB USING settings::JSONB;

-- Add indexes for performance
CREATE INDEX idx_widgets_user_id ON widgets(user_id);
CREATE INDEX idx_widgets_deleted_at ON widgets(deleted_at) WHERE deleted_at IS NULL;