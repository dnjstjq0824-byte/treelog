-- Migration: 0012_post_protection.up.sql
-- Purpose: Password-protected (PROTECTED) post visibility (G-07).
-- Adds a nullable Argon2id hash column and extends the visibility CHECK.

ALTER TABLE posts
    ADD COLUMN protection_password_hash TEXT;

ALTER TABLE posts
    DROP CONSTRAINT posts_visibility_check;

ALTER TABLE posts
    ADD CONSTRAINT posts_visibility_check
    CHECK (visibility IN ('PUBLIC', 'UNLISTED', 'NEIGHBOR_ONLY', 'MUTUAL_ONLY', 'PAID_MEMBERS_ONLY', 'PRIVATE', 'PROTECTED'));

-- A protected post without a password hash can never be unlocked; fail
-- closed at read time. Existing rows are unaffected (hash stays NULL and
-- no existing row uses the new visibility value).
COMMENT ON COLUMN posts.protection_password_hash IS 'Argon2id hash of the post protection password; NULL means no password set';
