-- Migration: 0012_post_protection.down.sql
-- Purpose: Revert 0012 (G-07). Only applicable when no post uses PROTECTED.

-- Guard: refuse to downgrade while protected posts exist.
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM posts WHERE visibility = 'PROTECTED') THEN
        RAISE EXCEPTION 'cannot downgrade 0012 while PROTECTED posts exist';
    END IF;
END
$$;

ALTER TABLE posts
    DROP CONSTRAINT posts_visibility_check;

ALTER TABLE posts
    ADD CONSTRAINT posts_visibility_check
    CHECK (visibility IN ('PUBLIC', 'UNLISTED', 'NEIGHBOR_ONLY', 'MUTUAL_ONLY', 'PAID_MEMBERS_ONLY', 'PRIVATE'));

ALTER TABLE posts
    DROP COLUMN protection_password_hash;
