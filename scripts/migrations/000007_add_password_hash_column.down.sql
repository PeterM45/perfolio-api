-- Remove password_hash column from users table
ALTER TABLE users DROP COLUMN IF EXISTS password_hash;

-- Note: We can't easily remove enum values in PostgreSQL without recreating the type
-- If you need to fully reverse this migration, you would need to:
-- 1. Create a new enum type without 'custom'
-- 2. Update all columns to use the new type
-- 3. Drop the old type
-- This is complex and risky, so we're not including it in the down migration