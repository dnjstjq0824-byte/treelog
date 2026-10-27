-- Migration: 0005_community.down.sql
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS notification_preferences;
DROP TABLE IF EXISTS user_blocks;
DROP TABLE IF EXISTS guestbook_entries;
DROP TABLE IF EXISTS bookmarks;
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS reactions;
DROP TABLE IF EXISTS subscriptions;
DROP TABLE IF EXISTS follows;
DROP TABLE IF EXISTS neighbor_groups;
