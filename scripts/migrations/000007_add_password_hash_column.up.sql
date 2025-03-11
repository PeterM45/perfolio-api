-- Add password_hash column to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash TEXT;

-- Add auth_provider_custom constant if needed
-- This assumes you have an enum type for auth_provider
DO $$
BEGIN
    -- Check if the enum type exists and if it already has the value
    IF EXISTS (
        SELECT 1 FROM pg_type WHERE typname = 'auth_provider'
    ) AND NOT EXISTS (
        SELECT 1 FROM pg_enum 
        WHERE enumtypid = (SELECT oid FROM pg_type WHERE typname = 'auth_provider')
        AND enumlabel = 'custom'
    ) THEN
        -- Add the new value to the enum
        ALTER TYPE auth_provider ADD VALUE 'custom';
    END IF;
EXCEPTION
    WHEN others THEN
        -- If anything goes wrong, we'll just continue
        RAISE NOTICE 'Could not add custom to auth_provider enum: %', SQLERRM;
END $$;