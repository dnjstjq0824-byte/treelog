-- Migration: 0007_moderation.down.sql
DROP TABLE IF EXISTS spam_detection_results;
DROP TABLE IF EXISTS appeals;
DROP TABLE IF EXISTS moderation_actions;
DROP TABLE IF EXISTS user_reports;
DROP TABLE IF EXISTS moderation_cases;
