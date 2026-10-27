-- Migration: 0003_content.down.sql
DROP TABLE IF EXISTS post_slug_history;
DROP TABLE IF EXISTS scheduled_publications;
DROP TABLE IF EXISTS post_visibility_rules;
DROP TABLE IF EXISTS post_tags;
DROP TABLE IF EXISTS tags;
DROP TABLE IF EXISTS post_categories;
DROP TABLE IF EXISTS autosave_drafts;
DROP TABLE IF EXISTS post_revisions;
DROP TABLE IF EXISTS posts;
DROP TABLE IF EXISTS series;
