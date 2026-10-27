-- Migration: 0008_monetization.up.sql
-- Domain: Monetization, Advertising, Memberships & Double-Entry Ledger (9 Tables: ad_accounts, ad_campaigns, ad_placements, ad_events, memberships, orders, payments, payouts, ledger_entries)

-- 52. ad_accounts
CREATE TABLE ad_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    company_name VARCHAR(160) NOT NULL,
    business_reg_no VARCHAR(64),
    billing_email CITEXT NOT NULL,
    currency VARCHAR(8) NOT NULL DEFAULT 'KRW',
    prepaid_balance_minor BIGINT NOT NULL DEFAULT 0 CHECK (prepaid_balance_minor >= 0),
    status VARCHAR(24) NOT NULL DEFAULT 'ACTIVE'
        CHECK (status IN ('PENDING_VERIFICATION', 'ACTIVE', 'SUSPENDED', 'CLOSED')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 53. ad_campaigns
CREATE TABLE ad_campaigns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ad_account_id UUID NOT NULL REFERENCES ad_accounts(id) ON DELETE RESTRICT,
    name VARCHAR(160) NOT NULL,
    pricing_model VARCHAR(16) NOT NULL CHECK (pricing_model IN ('CPM', 'CPC')),
    bid_amount_minor BIGINT NOT NULL CHECK (bid_amount_minor > 0),
    daily_budget_minor BIGINT NOT NULL CHECK (daily_budget_minor > 0),
    total_budget_minor BIGINT NOT NULL CHECK (total_budget_minor > 0),
    spent_amount_minor BIGINT NOT NULL DEFAULT 0 CHECK (spent_amount_minor >= 0),
    target_topics_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    target_languages_json JSONB NOT NULL DEFAULT '["ko"]'::jsonb,
    creative_title VARCHAR(120) NOT NULL,
    creative_body VARCHAR(300) NOT NULL,
    creative_media_id UUID REFERENCES media_assets(id) ON DELETE SET NULL,
    landing_url VARCHAR(1024) NOT NULL,
    review_status VARCHAR(24) NOT NULL DEFAULT 'PENDING_REVIEW'
        CHECK (review_status IN ('PENDING_REVIEW', 'APPROVED', 'REJECTED')),
    review_reject_reason VARCHAR(500),
    status VARCHAR(24) NOT NULL DEFAULT 'DRAFT'
        CHECK (status IN ('DRAFT', 'ACTIVE', 'PAUSED', 'BUDGET_EXHAUSTED', 'ENDED')),
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 54. ad_placements (Per-Blog Ad Slot Configuration)
CREATE TABLE ad_placements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    blog_id UUID NOT NULL REFERENCES blogs(id) ON DELETE CASCADE,
    slot_position VARCHAR(32) NOT NULL
        CHECK (slot_position IN ('POST_TOP', 'POST_MIDDLE', 'POST_BOTTOM', 'SIDEBAR_STICKY', 'LIST_INFEED')),
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    blocked_categories_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    revenue_share_bps INTEGER NOT NULL DEFAULT 7000 CHECK (revenue_share_bps BETWEEN 0 AND 10000),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (blog_id, slot_position)
);

-- 55. ad_events (Impressions & Clicks - Partitioned by created_at)
CREATE TABLE ad_events (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL REFERENCES ad_campaigns(id) ON DELETE RESTRICT,
    placement_id UUID NOT NULL REFERENCES ad_placements(id) ON DELETE RESTRICT,
    blog_id UUID NOT NULL REFERENCES blogs(id) ON DELETE RESTRICT,
    post_id UUID REFERENCES posts(id) ON DELETE SET NULL,
    event_type VARCHAR(16) NOT NULL CHECK (event_type IN ('IMPRESSION', 'CLICK')),
    cost_minor BIGINT NOT NULL DEFAULT 0 CHECK (cost_minor >= 0),
    creator_share_minor BIGINT NOT NULL DEFAULT 0 CHECK (creator_share_minor >= 0),
    platform_fee_minor BIGINT NOT NULL DEFAULT 0 CHECK (platform_fee_minor >= 0),
    visitor_hash CHAR(64) NOT NULL,
    is_invalid_traffic BOOLEAN NOT NULL DEFAULT FALSE,
    ivt_reason VARCHAR(64),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

CREATE TABLE ad_events_2026_10 PARTITION OF ad_events
    FOR VALUES FROM ('2026-10-01 00:00:00+00') TO ('2026-11-01 00:00:00+00');
CREATE TABLE ad_events_default PARTITION OF ad_events DEFAULT;

-- 56. memberships (Creator Paid Membership Tiers)
CREATE TABLE memberships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    blog_id UUID NOT NULL REFERENCES blogs(id) ON DELETE CASCADE,
    name VARCHAR(80) NOT NULL,
    description VARCHAR(500) NOT NULL DEFAULT '',
    monthly_price_minor BIGINT NOT NULL CHECK (monthly_price_minor > 0),
    annual_price_minor BIGINT CHECK (annual_price_minor IS NULL OR annual_price_minor > 0),
    currency VARCHAR(8) NOT NULL DEFAULT 'KRW',
    benefits_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 57. orders (Paid Post Single Purchase, Membership Subscription, Reader Donation, Ad Deposit)
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    idempotency_key VARCHAR(128) NOT NULL UNIQUE,
    buyer_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    recipient_blog_id UUID REFERENCES blogs(id) ON DELETE RESTRICT,
    order_type VARCHAR(32) NOT NULL
        CHECK (order_type IN ('MEMBERSHIP_SUB', 'SINGLE_POST_UNLOCK', 'READER_DONATION', 'AD_PREPAID_TOPUP')),
    target_ref_id UUID,
    gross_amount_minor BIGINT NOT NULL CHECK (gross_amount_minor > 0),
    platform_fee_minor BIGINT NOT NULL DEFAULT 0 CHECK (platform_fee_minor >= 0),
    net_creator_amount_minor BIGINT NOT NULL DEFAULT 0 CHECK (net_creator_amount_minor >= 0),
    currency VARCHAR(8) NOT NULL DEFAULT 'KRW',
    status VARCHAR(24) NOT NULL DEFAULT 'PENDING'
        CHECK (status IN ('PENDING', 'PAID', 'CANCELED', 'REFUNDED', 'DISPUTED')),
    access_expires_at TIMESTAMPTZ,
    donation_message VARCHAR(300),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 58. payments
CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE RESTRICT,
    pg_provider VARCHAR(32) NOT NULL CHECK (pg_provider IN ('STRIPE', 'TOSS_PAYMENTS', 'PORTONE', 'INTERNAL_MOCK')),
    pg_transaction_id VARCHAR(255) NOT NULL UNIQUE,
    payment_method VARCHAR(32) NOT NULL,
    amount_minor BIGINT NOT NULL CHECK (amount_minor > 0),
    currency VARCHAR(8) NOT NULL DEFAULT 'KRW',
    status VARCHAR(24) NOT NULL
        CHECK (status IN ('AUTHORIZED', 'CAPTURED', 'FAILED', 'REFUNDED', 'CHARGEBACK')),
    paid_at TIMESTAMPTZ,
    refunded_at TIMESTAMPTZ,
    raw_webhook_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 59. payouts (Creator Revenue Withdrawal & Settlement)
CREATE TABLE payouts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    idempotency_key VARCHAR(128) NOT NULL UNIQUE,
    payee_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    blog_id UUID NOT NULL REFERENCES blogs(id) ON DELETE RESTRICT,
    gross_payout_minor BIGINT NOT NULL CHECK (gross_payout_minor > 0),
    withholding_tax_minor BIGINT NOT NULL DEFAULT 0 CHECK (withholding_tax_minor >= 0),
    net_payout_minor BIGINT NOT NULL CHECK (net_payout_minor > 0),
    currency VARCHAR(8) NOT NULL DEFAULT 'KRW',
    bank_code VARCHAR(32) NOT NULL,
    bank_account_enc BYTEA NOT NULL,
    status VARCHAR(24) NOT NULL DEFAULT 'REQUESTED'
        CHECK (status IN ('REQUESTED', 'ON_HOLD', 'APPROVED', 'PROCESSING', 'PAID', 'REJECTED', 'CLAWED_BACK')),
    hold_reason VARCHAR(500),
    approved_by UUID REFERENCES users(id) ON DELETE SET NULL,
    paid_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 60. ledger_entries (Immutable Double-Entry Accounting Ledger)
CREATE TABLE ledger_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_group_id UUID NOT NULL,
    idempotency_key VARCHAR(160) NOT NULL UNIQUE,
    account_type VARCHAR(32) NOT NULL
        CHECK (account_type IN ('PLATFORM_CLEARING', 'ADVERTISER_PREPAID', 'CREATOR_PENDING_ESCROW', 'CREATOR_PAYABLE', 'PLATFORM_FEE_REVENUE', 'TAX_WITHHOLDING', 'PAYOUT_DISBURSED')),
    owner_user_id UUID REFERENCES users(id) ON DELETE RESTRICT,
    blog_id UUID REFERENCES blogs(id) ON DELETE RESTRICT,
    direction VARCHAR(8) NOT NULL CHECK (direction IN ('DEBIT', 'CREDIT')),
    amount_minor BIGINT NOT NULL CHECK (amount_minor > 0),
    currency VARCHAR(8) NOT NULL DEFAULT 'KRW',
    reference_type VARCHAR(32) NOT NULL
        CHECK (reference_type IN ('AD_DAILY_SETTLEMENT', 'ORDER_PAYMENT', 'ORDER_REFUND', 'PAYOUT_SETTLEMENT', 'FRAUD_CLAWBACK')),
    reference_id UUID NOT NULL,
    memo VARCHAR(300) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ledger_owner_account ON ledger_entries(owner_user_id, account_type, created_at DESC);
CREATE INDEX idx_orders_buyer_target ON orders(buyer_user_id, target_ref_id, status);
