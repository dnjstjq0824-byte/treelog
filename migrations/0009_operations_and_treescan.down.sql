-- Migration: 0009_operations_and_treescan.down.sql
DROP TABLE IF EXISTS treescan_sync_state;
DROP TABLE IF EXISTS migration_jobs;
DROP TABLE IF EXISTS background_jobs;
DROP TABLE IF EXISTS outbox_events;
DROP TABLE IF EXISTS system_settings;
DROP TABLE IF EXISTS feature_flags;
DROP TABLE IF EXISTS audit_logs;
