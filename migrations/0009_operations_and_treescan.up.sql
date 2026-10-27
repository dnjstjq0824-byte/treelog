-- Migration: 0009_operations_and_treescan.up.sql
-- Domain: System Operations, Audit Logs, Portability & TreeScan Integration Readiness (7 Tables: audit_logs, feature_flags, system_settings, outbox_events, background_jobs, migration_jobs, treescan_sync_state)

-- 61. audit_logs (Immutable Security & Admin Audit Trail - Partitioned by created_at)
CREATE TABLE audit_logs (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    actor_user_id UUID REFERENCES users(id) ON DELETE RESTRICT,
    actor_role VARCHAR(40) NOT NULL,
    action_code VARCHAR(80) NOT NULL,
    target_type VARCHAR(40) NOT NULL,
    target_id VARCHAR(128) NOT NULL,
    before_state_json JSONB,
    after_state_json JSONB,
    reason VARCHAR(500) NOT NULL DEFAULT '',
    ip_address INET,
    user_agent VARCHAR(512),
    request_id VARCHAR(64),
    prev_hash CHAR(64),
    entry_hash CHAR(64) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

CREATE TABLE audit_logs_2026_10 PARTITION OF audit_logs
    FOR VALUES FROM ('2026-10-01 00:00:00+00') TO ('2026-11-01 00:00:00+00');
CREATE TABLE audit_logs_default PARTITION OF audit_logs DEFAULT;

-- 62. feature_flags
CREATE TABLE feature_flags (
    key VARCHAR(80) PRIMARY KEY,
    description VARCHAR(300) NOT NULL,
    is_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    rollout_percentage SMALLINT NOT NULL DEFAULT 0 CHECK (rollout_percentage BETWEEN 0 AND 100),
    allowlist_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    updated_by UUID REFERENCES users(id) ON DELETE SET NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 63. system_settings
CREATE TABLE system_settings (
    key VARCHAR(80) PRIMARY KEY,
    value_json JSONB NOT NULL,
    description VARCHAR(300) NOT NULL DEFAULT '',
    is_secret BOOLEAN NOT NULL DEFAULT FALSE,
    updated_by UUID REFERENCES users(id) ON DELETE SET NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 64. outbox_events (Transactional Outbox Pattern for Async Event Delivery)
CREATE TABLE outbox_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_type VARCHAR(40) NOT NULL,
    aggregate_id UUID NOT NULL,
    event_type VARCHAR(80) NOT NULL,
    payload_json JSONB NOT NULL,
    idempotency_key VARCHAR(160) NOT NULL UNIQUE,
    status VARCHAR(24) NOT NULL DEFAULT 'PENDING'
        CHECK (status IN ('PENDING', 'PROCESSING', 'PUBLISHED', 'DEAD_LETTER')),
    retry_count SMALLINT NOT NULL DEFAULT 0,
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at TIMESTAMPTZ
);

-- 65. background_jobs
CREATE TABLE background_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    queue_name VARCHAR(32) NOT NULL CHECK (queue_name IN ('critical', 'default', 'media', 'bulk')),
    job_type VARCHAR(64) NOT NULL,
    payload_json JSONB NOT NULL,
    idempotency_key VARCHAR(160) UNIQUE,
    status VARCHAR(24) NOT NULL DEFAULT 'QUEUED'
        CHECK (status IN ('QUEUED', 'EXECUTING', 'COMPLETED', 'FAILED', 'DEAD_LETTER')),
    attempt_count SMALLINT NOT NULL DEFAULT 0,
    max_attempts SMALLINT NOT NULL DEFAULT 8,
    last_error TEXT,
    scheduled_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 66. migration_jobs (External Blog Import & Content Backup Export Jobs)
CREATE TABLE migration_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    blog_id UUID NOT NULL REFERENCES blogs(id) ON DELETE CASCADE,
    initiated_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    direction VARCHAR(16) NOT NULL CHECK (direction IN ('IMPORT', 'EXPORT')),
    format VARCHAR(24) NOT NULL CHECK (format IN ('MARKDOWN_ZIP', 'HTML_ZIP', 'WXR_XML', 'RSS_XML')),
    status VARCHAR(24) NOT NULL DEFAULT 'QUEUED'
        CHECK (status IN ('QUEUED', 'PROCESSING', 'COMPLETED', 'PARTIAL_SUCCESS', 'FAILED')),
    total_items INTEGER NOT NULL DEFAULT 0,
    processed_items INTEGER NOT NULL DEFAULT 0,
    failed_items INTEGER NOT NULL DEFAULT 0,
    archive_media_id UUID REFERENCES media_assets(id) ON DELETE SET NULL,
    error_report_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    ownership_confirmed BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

-- 67. treescan_sync_state (Future TreeScan Search Engine Sync Cursor & Tombstone Log)
CREATE TABLE treescan_sync_state (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id UUID NOT NULL UNIQUE REFERENCES posts(id) ON DELETE CASCADE,
    blog_id UUID NOT NULL REFERENCES blogs(id) ON DELETE CASCADE,
    sync_action VARCHAR(16) NOT NULL CHECK (sync_action IN ('UPSERT', 'DELETE_TOMBSTONE')),
    visibility_snapshot VARCHAR(24) NOT NULL,
    payload_hash CHAR(64) NOT NULL,
    sync_status VARCHAR(24) NOT NULL DEFAULT 'PENDING'
        CHECK (sync_status IN ('PENDING', 'SYNCED', 'SKIPPED_NOT_PUBLIC', 'FAILED')),
    sequence_cursor BIGSERIAL NOT NULL,
    last_synced_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_outbox_pending ON outbox_events(next_attempt_at ASC) WHERE status = 'PENDING';
CREATE INDEX idx_treescan_sync_cursor ON treescan_sync_state(sequence_cursor ASC) WHERE sync_status = 'PENDING';
