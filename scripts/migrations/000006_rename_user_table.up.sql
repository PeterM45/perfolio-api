-- Rename the table from 'user' to 'users' to match convention
ALTER TABLE IF EXISTS "user" RENAME TO users;

-- Update any sequence names if they exist
ALTER SEQUENCE IF EXISTS user_id_seq RENAME TO users_id_seq;

-- You may need to update foreign key constraints if they reference this table
-- Examples (update based on your actual constraints):
-- ALTER TABLE follows RENAME CONSTRAINT follows_follower_id_fkey TO follows_users_follower_id_fkey;
-- ALTER TABLE follows RENAME CONSTRAINT follows_following_id_fkey TO follows_users_following_id_fkey;
