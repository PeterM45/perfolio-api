-- Rename the table back from 'users' to 'user'
ALTER TABLE IF EXISTS users RENAME TO "user";

-- Revert any sequence names if changed
ALTER SEQUENCE IF EXISTS users_id_seq RENAME TO user_id_seq;

-- Revert foreign key constraints if they were updated
-- Examples (update based on your actual constraints):
-- ALTER TABLE follows RENAME CONSTRAINT follows_users_follower_id_fkey TO follows_follower_id_fkey;
-- ALTER TABLE follows RENAME CONSTRAINT follows_users_following_id_fkey TO follows_following_id_fkey;
