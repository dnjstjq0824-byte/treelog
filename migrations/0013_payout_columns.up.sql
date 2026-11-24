-- Migration: 0013_payout_columns.up.sql
-- Purpose: Restore payout domain columns the service layer requires
-- (settlement period and recipient holder name). Additive only.

ALTER TABLE payouts
    ADD COLUMN period_start TIMESTAMPTZ,
    ADD COLUMN period_end TIMESTAMPTZ,
    ADD COLUMN account_holder VARCHAR(120);

COMMENT ON COLUMN payouts.period_start IS 'Settlement period covered by this payout (inclusive start)';
COMMENT ON COLUMN payouts.period_end IS 'Settlement period covered by this payout (inclusive end)';
COMMENT ON COLUMN payouts.account_holder IS 'Recipient account holder name for the bank transfer';
