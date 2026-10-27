-- Migration: 0008_monetization.down.sql
DROP TABLE IF EXISTS ledger_entries;
DROP TABLE IF EXISTS payouts;
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS memberships;
DROP TABLE IF EXISTS ad_events;
DROP TABLE IF EXISTS ad_placements;
DROP TABLE IF EXISTS ad_campaigns;
DROP TABLE IF EXISTS ad_accounts;
