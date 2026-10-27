-- Migration: 0002_blog.up.sql
-- Domain: Blog & Domains & Themes (8 Tables: blogs, blog_members, blog_settings, blog_domains, themes, blog_themes, blog_navigation, blog_categories)

-- 10. blogs
CREATE TABLE blogs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    handle CITEXT NOT NULL UNIQUE CHECK (handle ~ '^[a-z0-9][a-z0-9-]{1,28}[a-z0-9]$'),
    blog_type VARCHAR(24) NOT NULL DEFAULT 'PERSONAL'
        CHECK (blog_type IN ('PERSONAL', 'PROFESSIONAL', 'CORPORATE', 'COMMUNITY')),
    title VARCHAR(120) NOT NULL,
    subtitle VARCHAR(240) NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    logo_media_id UUID,
    cover_media_id UUID,
    favicon_media_id UUID,
    default_language VARCHAR(16) NOT NULL DEFAULT 'ko-KR',
    timezone VARCHAR(64) NOT NULL DEFAULT 'Asia/Seoul',
    primary_url_type VARCHAR(24) NOT NULL DEFAULT 'SUBDOMAIN'
        CHECK (primary_url_type IN ('SUBDOMAIN', 'PATH', 'CUSTOM_DOMAIN')),
    status VARCHAR(24) NOT NULL DEFAULT 'ACTIVE'
        CHECK (status IN ('ACTIVE', 'READ_ONLY', 'SUSPENDED', 'CLOSING', 'CLOSED', 'PURGED')),
    subscriber_count BIGINT NOT NULL DEFAULT 0 CHECK (subscriber_count >= 0),
    neighbor_count BIGINT NOT NULL DEFAULT 0 CHECK (neighbor_count >= 0),
    post_count INTEGER NOT NULL DEFAULT 0 CHECK (post_count >= 0),
    closing_requested_at TIMESTAMPTZ,
    closed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 11. blog_members
CREATE TABLE blog_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    blog_id UUID NOT NULL REFERENCES blogs(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(24) NOT NULL
        CHECK (role IN ('OWNER', 'ADMIN', 'EDITOR', 'AUTHOR', 'ANALYST')),
    invite_status VARCHAR(24) NOT NULL DEFAULT 'ACCEPTED'
        CHECK (invite_status IN ('PENDING', 'ACCEPTED', 'DECLINED', 'REVOKED')),
    invited_by UUID REFERENCES users(id) ON DELETE SET NULL,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (blog_id, user_id)
);

-- 12. blog_settings
CREATE TABLE blog_settings (
    blog_id UUID PRIMARY KEY REFERENCES blogs(id) ON DELETE CASCADE,
    default_post_visibility VARCHAR(24) NOT NULL DEFAULT 'PUBLIC'
        CHECK (default_post_visibility IN ('PUBLIC', 'UNLISTED', 'NEIGHBOR_ONLY', 'MUTUAL_ONLY', 'PAID_MEMBERS_ONLY', 'PRIVATE')),
    default_allow_comments VARCHAR(24) NOT NULL DEFAULT 'ALL_ALLOWED'
        CHECK (default_allow_comments IN ('ALL_ALLOWED', 'NEIGHBORS_ONLY', 'MUTUAL_ONLY', 'REQUIRE_APPROVAL', 'DISABLED')),
    allow_guestbook BOOLEAN NOT NULL DEFAULT TRUE,
    allow_mutual_requests BOOLEAN NOT NULL DEFAULT TRUE,
    allow_rss_full_content BOOLEAN NOT NULL DEFAULT TRUE,
    allow_search_indexing BOOLEAN NOT NULL DEFAULT TRUE,
    allow_treescan_sync BOOLEAN NOT NULL DEFAULT TRUE,
    anti_copy_protection BOOLEAN NOT NULL DEFAULT FALSE,
    require_approval_for_authors BOOLEAN NOT NULL DEFAULT FALSE,
    posts_per_page SMALLINT NOT NULL DEFAULT 10 CHECK (posts_per_page BETWEEN 5 AND 50),
    show_visitor_counter BOOLEAN NOT NULL DEFAULT TRUE,
    adult_content_warning BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 13. blog_domains
CREATE TABLE blog_domains (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    blog_id UUID NOT NULL REFERENCES blogs(id) ON DELETE CASCADE,
    domain_name CITEXT NOT NULL UNIQUE,
    domain_type VARCHAR(24) NOT NULL
        CHECK (domain_type IN ('SUBDOMAIN', 'CUSTOM_CNAME', 'CUSTOM_APEX')),
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    verification_token VARCHAR(128) NOT NULL,
    verification_status VARCHAR(24) NOT NULL DEFAULT 'PENDING_DNS'
        CHECK (verification_status IN ('PENDING_DNS', 'VERIFIED', 'FAILED', 'EXPIRED', 'SUSPENDED')),
    tls_status VARCHAR(24) NOT NULL DEFAULT 'NONE'
        CHECK (tls_status IN ('NONE', 'PROVISIONING', 'ACTIVE', 'RENEWAL_FAILED')),
    tls_expires_at TIMESTAMPTZ,
    redirect_to_primary BOOLEAN NOT NULL DEFAULT TRUE,
    last_verified_at TIMESTAMPTZ,
    failure_reason VARCHAR(500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_blog_domains_one_primary ON blog_domains(blog_id) WHERE is_primary = TRUE;

-- 14. themes (Platform Official Theme Catalog)
CREATE TABLE themes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug VARCHAR(64) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    description VARCHAR(500) NOT NULL DEFAULT '',
    layout_type VARCHAR(32) NOT NULL
        CHECK (layout_type IN ('MAGAZINE', 'MINIMAL', 'TECH_DOCS', 'PHOTO_GALLERY', 'CLASSIC_BLOG', 'EDITORIAL')),
    default_tokens_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    preview_image_url VARCHAR(512),
    version VARCHAR(24) NOT NULL DEFAULT '1.0.0',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 15. blog_themes (Per-Blog Applied Theme & Customization)
CREATE TABLE blog_themes (
    blog_id UUID PRIMARY KEY REFERENCES blogs(id) ON DELETE CASCADE,
    theme_id UUID NOT NULL REFERENCES themes(id) ON DELETE RESTRICT,
    color_mode VARCHAR(16) NOT NULL DEFAULT 'SYSTEM'
        CHECK (color_mode IN ('LIGHT', 'DARK', 'SYSTEM')),
    tokens_override_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    layout_config_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    sidebar_widgets_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    custom_css_sanitized TEXT NOT NULL DEFAULT '',
    footer_text VARCHAR(500) NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 16. blog_navigation
CREATE TABLE blog_navigation (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    blog_id UUID NOT NULL REFERENCES blogs(id) ON DELETE CASCADE,
    label VARCHAR(80) NOT NULL,
    link_type VARCHAR(24) NOT NULL
        CHECK (link_type IN ('CATEGORY', 'SERIES', 'PAGE', 'EXTERNAL_URL', 'GUESTBOOK', 'TAGS')),
    target_ref VARCHAR(512) NOT NULL,
    open_in_new_tab BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order SMALLINT NOT NULL DEFAULT 0,
    is_visible BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 17. blog_categories (Up to 2-depth hierarchy)
CREATE TABLE blog_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    blog_id UUID NOT NULL REFERENCES blogs(id) ON DELETE CASCADE,
    parent_id UUID REFERENCES blog_categories(id) ON DELETE SET NULL,
    name VARCHAR(80) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    description VARCHAR(300) NOT NULL DEFAULT '',
    topic_code VARCHAR(40),
    list_style VARCHAR(24) NOT NULL DEFAULT 'CARD'
        CHECK (list_style IN ('CARD', 'LIST', 'GALLERY', 'COMPACT')),
    visibility VARCHAR(24) NOT NULL DEFAULT 'PUBLIC'
        CHECK (visibility IN ('PUBLIC', 'NEIGHBOR_ONLY', 'MUTUAL_ONLY', 'PRIVATE')),
    sort_order INTEGER NOT NULL DEFAULT 0,
    post_count INTEGER NOT NULL DEFAULT 0 CHECK (post_count >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (blog_id, slug)
);

CREATE INDEX idx_blogs_owner ON blogs(owner_user_id);
CREATE INDEX idx_blog_categories_blog_sort ON blog_categories(blog_id, sort_order);
