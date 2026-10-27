-- Migration: 0006_search.up.sql
-- Domain: OpenSearch & Content Quality Signals (4 Tables: search_index_jobs, search_queries, search_click_events, content_quality_signals)

-- 43. search_index_jobs
CREATE TABLE search_index_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    target_index VARCHAR(40) NOT NULL
        CHECK (target_index IN ('POSTS', 'BLOGS', 'SUGGEST')),
    document_id UUID NOT NULL,
    operation VARCHAR(16) NOT NULL
        CHECK (operation IN ('UPSERT', 'DELETE')),
    priority SMALLINT NOT NULL DEFAULT 50 CHECK (priority BETWEEN 1 AND 100),
    status VARCHAR(24) NOT NULL DEFAULT 'PENDING'
        CHECK (status IN ('PENDING', 'PROCESSING', 'COMPLETED', 'FAILED', 'DEAD_LETTER')),
    attempt_count SMALLINT NOT NULL DEFAULT 0,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ
);

-- 44. search_queries
CREATE TABLE search_queries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    normalized_query VARCHAR(200) NOT NULL,
    language_code VARCHAR(16) NOT NULL DEFAULT 'ko',
    blog_scope_id UUID REFERENCES blogs(id) ON DELETE SET NULL,
    result_count INTEGER NOT NULL DEFAULT 0 CHECK (result_count >= 0),
    is_blocked_spam_term BOOLEAN NOT NULL DEFAULT FALSE,
    last_searched_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    search_count BIGINT NOT NULL DEFAULT 1 CHECK (search_count >= 1),
    UNIQUE (normalized_query, language_code)
);

-- 45. search_click_events (Partitioned by created_at)
CREATE TABLE search_click_events (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    query_text VARCHAR(200) NOT NULL,
    clicked_post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    rank_position SMALLINT NOT NULL CHECK (rank_position >= 1),
    session_hash CHAR(64) NOT NULL,
    dwell_time_sec INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

CREATE TABLE search_click_events_2026_10 PARTITION OF search_click_events
    FOR VALUES FROM ('2026-10-01 00:00:00+00') TO ('2026-11-01 00:00:00+00');
CREATE TABLE search_click_events_default PARTITION OF search_click_events DEFAULT;

-- 46. content_quality_signals
CREATE TABLE content_quality_signals (
    post_id UUID PRIMARY KEY REFERENCES posts(id) ON DELETE CASCADE,
    simhash_64 BIGINT NOT NULL DEFAULT 0,
    word_count INTEGER NOT NULL DEFAULT 0,
    image_count INTEGER NOT NULL DEFAULT 0,
    heading_count INTEGER NOT NULL DEFAULT 0,
    external_link_count INTEGER NOT NULL DEFAULT 0,
    quality_score NUMERIC(5,2) NOT NULL DEFAULT 50.00 CHECK (quality_score BETWEEN 0 AND 100),
    spam_probability NUMERIC(4,3) NOT NULL DEFAULT 0.000 CHECK (spam_probability BETWEEN 0 AND 1),
    is_near_duplicate BOOLEAN NOT NULL DEFAULT FALSE,
    original_source_post_id UUID REFERENCES posts(id) ON DELETE SET NULL,
    index_eligibility VARCHAR(32) NOT NULL DEFAULT 'ELIGIBLE'
        CHECK (index_eligibility IN ('ELIGIBLE', 'NOINDEX_LOW_QUALITY', 'NOINDEX_DUPLICATE', 'QUARANTINED_SPAM')),
    computed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_search_jobs_pending ON search_index_jobs(priority DESC, created_at ASC) WHERE status = 'PENDING';
CREATE INDEX idx_quality_simhash ON content_quality_signals(simhash_64);
