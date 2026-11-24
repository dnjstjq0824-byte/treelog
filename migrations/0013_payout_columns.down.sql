-- Migration: 0013_payout_columns.down.sql
-- Purpose: Revert 0013. Drops period/holder columns (data loss on downgrade).

ALTER TABLE payouts
    DROP COLUMN period_start,
    DROP COLUMN period_end,
    DROP COLUMN account_holder;
