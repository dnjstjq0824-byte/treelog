-- Migration: 0003_content.up.sql
-- Domain: Content & Editor Engine (11 Tables: series, posts, post_revisions, autosave_drafts, post_categories, tags, post_tags, post_media, post_visibility_rules, scheduled_publications, post_slug_history)

-- 18. series
CREATE TABLE series (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    blog_id UUID NOT NULL REFERENCES blogs(id) ON DELETE CASCADE,
    title VARCHAR(160) NOT NULL,
    slug VARCHAR(180) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    cover_media_id UUID,
    is_completed BOOLEAN NOT NULL DEFAULT FALSE,
    post_count INTEGER NOT NULL DEFAULT 0 CHECK (post_count >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (blog_id, slug)
);

-- 19. posts
CREATE TABLE posts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    blog_id UUID NOT NULL REFERENCES blogs(id) ON DELETE CASCADE,
    author_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    category_id UUID REFERENCES blog_categories(id) ON DELETE SET NULL,
    series_id UUID REFERENCES series(id) ON DELETE SET NULL,
    series_order INTEGER,
    slug VARCHAR(220) NOT NULL,
    title VARCHAR(300) NOT NULL,
    subtitle VARCHAR(500) NOT NULL DEFAULT '',
    content_json JSONB NOT NULL DEFAULT '{"type":"doc","content":[]}'::jsonb,
    content_html TEXT NOT NULL DEFAULT '',
    content_plain TEXT NOT NULL DEFAULT '',
    excerpt VARCHAR(600) NOT NULL DEFAULT '',
    cover_media_id UUID,
    language_code VARCHAR(16) NOT NULL DEFAULT 'ko',
    translation_group_id UUID,
    status VARCHAR(24) NOT NULL DEFAULT 'DRAFT'
        CHECK (status IN ('DRAFT', 'IN_REVIEW', 'SCHEDULED', 'PUBLISHED', 'UNPUBLISHED', 'QUARANTINED', 'TRASHED', 'PURGED')),
    visibility VARCHAR(24) NOT NULL DEFAULT 'PUBLIC'
        CHECK (visibility IN ('PUBLIC', 'UNLISTED', 'NEIGHBOR_ONLY', 'MUTUAL_ONLY', 'PAID_MEMBERS_ONLY', 'PRIVATE')),
    allow_comments VARCHAR(24) NOT NULL DEFAULT 'ALL_ALLOWED'
        CHECK (allow_comments IN ('ALL_ALLOWED', 'NEIGHBORS_ONLY', 'MUTUAL_ONLY', 'REQUIRE_APPROVAL', 'DISABLED')),
    allow_reactions BOOLEAN NOT NULL DEFAULT TRUE,
    is_notice BOOLEAN NOT NULL DEFAULT FALSE,
    is_pinned BOOLEAN NOT NULL DEFAULT FALSE,
    seo_title VARCHAR(200),
    seo_description VARCHAR(320),
    canonical_url VARCHAR(1024),
    source_citation_urls JSONB NOT NULL DEFAULT '[]'::jsonb,
    license_type VARCHAR(32) NOT NULL DEFAULT 'ALL_RIGHTS_RESERVED'
        CHECK (license_type IN ('ALL_RIGHTS_RESERVED', 'CC_BY', 'CC_BY_NC', 'CC_BY_ND', 'CC_BY_SA', 'CC_BY_NC_SA', 'CC_BY_NC_ND')),
    ai_disclosure_level VARCHAR(24) NOT NULL DEFAULT 'NONE'
        CHECK (ai_disclosure_level IN ('NONE', 'AI_ASSISTED', 'AI_GENERATED')),
    paywall_price_minor BIGINT CHECK (paywall_price_minor IS NULL OR paywall_price_minor >= 0),
    word_count INTEGER NOT NULL DEFAULT 0 CHECK (word_count >= 0),
    reading_time_sec INTEGER NOT NULL DEFAULT 0 CHECK (reading_time_sec >= 0),
    view_count BIGINT NOT NULL DEFAULT 0 CHECK (view_count >= 0),
    reaction_count INTEGER NOT NULL DEFAULT 0 CHECK (reaction_count >= 0),
    comment_count INTEGER NOT NULL DEFAULT 0 CHECK (comment_count >= 0),
    bookmark_count INTEGER NOT NULL DEFAULT 0 CHECK (bookmark_count >= 0),
    revision_no INTEGER NOT NULL DEFAULT 1 CHECK (revision_no >= 1),
    published_at TIMESTAMPTZ,
    trashed_at TIMESTAMPTZ,
    purge_after_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (blog_id, slug)
);

-- 20. post_revisions
CREATE TABLE post_revisions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    revision_no INTEGER NOT NULL CHECK (revision_no >= 1),
    editor_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    title VARCHAR(300) NOT NULL,
    subtitle VARCHAR(500) NOT NULL DEFAULT '',
    content_json JSONB NOT NULL,
    content_html TEXT NOT NULL,
    change_summary VARCHAR(240) NOT NULL DEFAULT '',
    is_published_snapshot BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (post_id, revision_no)
);

-- 21. autosave_drafts
CREATE TABLE autosave_drafts (
    post_id UUID PRIMARY KEY REFERENCES posts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    base_revision_no INTEGER NOT NULL DEFAULT 1,
    title VARCHAR(300) NOT NULL DEFAULT '',
    subtitle VARCHAR(500) NOT NULL DEFAULT '',
    content_json JSONB NOT NULL,
    client_updated_at TIMESTAMPTZ NOT NULL,
    server_saved_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 22. post_categories (Secondary multi-category mapping if needed)
CREATE TABLE post_categories (
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES blog_categories(id) ON DELETE CASCADE,
    is_primary BOOLEAN NOT NULL DEFAULT TRUE,
    PRIMARY KEY (post_id, category_id)
);

-- 23. tags
CREATE TABLE tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name CITEXT NOT NULL UNIQUE,
    slug CITEXT NOT NULL UNIQUE,
    post_count BIGINT NOT NULL DEFAULT 0 CHECK (post_count >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 24. post_tags
CREATE TABLE post_tags (
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    sort_order SMALLINT NOT NULL DEFAULT 0,
    PRIMARY KEY (post_id, tag_id)
);

-- 25. post_visibility_rules (Fine-grained neighbor groups or paid membership tier rules)
CREATE TABLE post_visibility_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    rule_type VARCHAR(32) NOT NULL
        CHECK (rule_type IN ('NEIGHBOR_GROUP', 'MEMBERSHIP_TIER', 'PASSWORD_PROTECTED')),
    target_ref_id UUID,
    access_password_hash TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 26. scheduled_publications
CREATE TABLE scheduled_publications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id UUID NOT NULL UNIQUE REFERENCES posts(id) ON DELETE CASCADE,
    blog_id UUID NOT NULL REFERENCES blogs(id) ON DELETE CASCADE,
    scheduled_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    scheduled_for TIMESTAMPTZ NOT NULL,
    timezone VARCHAR(64) NOT NULL DEFAULT 'Asia/Seoul',
    status VARCHAR(24) NOT NULL DEFAULT 'PENDING'
        CHECK (status IN ('PENDING', 'EXECUTING', 'COMPLETED', 'CANCELED', 'FAILED')),
    executed_at TIMESTAMPTZ,
    failure_reason VARCHAR(500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 27. post_slug_history (For 301 Redirects on slug change or external blog import)
CREATE TABLE post_slug_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    blog_id UUID NOT NULL REFERENCES blogs(id) ON DELETE CASCADE,
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    old_slug VARCHAR(512) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (blog_id, old_slug)
);

CREATE INDEX idx_posts_blog_published ON posts(blog_id, published_at DESC, id DESC) WHERE status = 'PUBLISHED' AND trashed_at IS NULL;
CREATE INDEX idx_posts_public_discover ON posts(published_at DESC, id DESC) WHERE status = 'PUBLISHED' AND visibility = 'PUBLIC' AND trashed_at IS NULL;
CREATE INDEX idx_scheduled_pubs_pending ON scheduled_publications(scheduled_for) WHERE status = 'PENDING';
