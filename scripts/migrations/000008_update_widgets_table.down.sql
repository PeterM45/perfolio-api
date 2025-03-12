-- Remove indexes
DROP INDEX IF EXISTS idx_widgets_user_id;
DROP INDEX IF EXISTS idx_widgets_deleted_at;

-- Convert settings back to text
ALTER TABLE widgets ALTER COLUMN settings TYPE TEXT USING settings::TEXT;

-- Remove columns
ALTER TABLE widgets DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE widgets DROP COLUMN IF EXISTS display_name;
ALTER TABLE widgets DROP COLUMN IF EXISTS is_visible;
ALTER TABLE widgets DROP COLUMN IF EXISTS version;