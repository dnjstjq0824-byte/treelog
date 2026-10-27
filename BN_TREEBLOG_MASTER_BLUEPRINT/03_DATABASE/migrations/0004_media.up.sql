-- Migration: 0004_media.up.sql
-- Domain: Video & Media Engine (5 Tables: upload_sessions, media_assets, media_variants, media_processing_jobs, post_media)

-- 28. upload_sessions
CREATE TABLE upload_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    blog_id UUID NOT NULL REFERENCES blogs(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    original_filename VARCHAR(255) NOT NULL,
    mime_type VARCHAR(120) NOT NULL,
    expected_size_bytes BIGINT NOT NULL CHECK (expected_size_bytes > 0),
    sha256_hash CHAR(64),
    storage_bucket VARCHAR(120) NOT NULL,
    storage_object_key VARCHAR(512) NOT NULL UNIQUE,
    status VARCHAR(24) NOT NULL DEFAULT 'ISSUED'
        CHECK (status IN ('ISSUED', 'UPLOADED', 'COMPLETED', 'EXPIRED', 'ABORTED')),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 29. media_assets
CREATE TABLE media_assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    blog_id UUID NOT NULL REFERENCES blogs(id) ON DELETE CASCADE,
    uploader_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    upload_session_id UUID REFERENCES upload_sessions(id) ON DELETE SET NULL,
    media_type VARCHAR(24) NOT NULL
        CHECK (media_type IN ('IMAGE', 'ANIMATED_GIF', 'VIDEO', 'VIDEO_EMBED', 'ARCHIVE')),
    mime_type VARCHAR(120) NOT NULL,
    original_size_bytes BIGINT NOT NULL DEFAULT 0 CHECK (original_size_bytes >= 0),
    sha256_hash CHAR(64) NOT NULL,
    width INTEGER CHECK (width IS NULL OR width > 0),
    height INTEGER CHECK (height IS NULL OR height > 0),
    duration_sec NUMERIC(10,2),
    blurhash VARCHAR(100),
    alt_text VARCHAR(500) NOT NULL DEFAULT '',
    caption VARCHAR(500) NOT NULL DEFAULT '',
    private_storage_key VARCHAR(512),
    embed_provider VARCHAR(40),
    embed_external_id VARCHAR(255),
    embed_html_sanitized TEXT,
    exif_stripped BOOLEAN NOT NULL DEFAULT FALSE,
    virus_scan_status VARCHAR(24) NOT NULL DEFAULT 'PENDING'
        CHECK (virus_scan_status IN ('PENDING', 'CLEAN', 'INFECTED', 'SKIPPED')),
    status VARCHAR(24) NOT NULL DEFAULT 'PROCESSING'
        CHECK (status IN ('UPLOADING', 'PROCESSING', 'READY', 'QUARANTINED', 'DELETED')),
    reference_count INTEGER NOT NULL DEFAULT 0 CHECK (reference_count >= 0),
    orphan_since TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 30. media_variants
CREATE TABLE media_variants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    media_asset_id UUID NOT NULL REFERENCES media_assets(id) ON DELETE CASCADE,
    variant_name VARCHAR(32) NOT NULL
        CHECK (variant_name IN ('THUMB_400', 'W320', 'W640', 'W1200', 'W1920', 'ANIMATED_WEBP', 'LOOP_MP4', 'HLS_PLAYLIST')),
    format VARCHAR(16) NOT NULL
        CHECK (format IN ('WEBP', 'AVIF', 'JPEG', 'PNG', 'MP4', 'M3U8')),
    width INTEGER NOT NULL CHECK (width > 0),
    height INTEGER NOT NULL CHECK (height > 0),
    size_bytes BIGINT NOT NULL CHECK (size_bytes > 0),
    public_cdn_url VARCHAR(1024) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (media_asset_id, variant_name, format)
);

-- 31. media_processing_jobs
CREATE TABLE media_processing_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    media_asset_id UUID NOT NULL REFERENCES media_assets(id) ON DELETE CASCADE,
    job_type VARCHAR(32) NOT NULL
        CHECK (job_type IN ('IMAGE_OPTIMIZE_AND_STRIP_EXIF', 'GIF_TO_MP4', 'VIDEO_TRANSCODE_HLS')),
    status VARCHAR(24) NOT NULL DEFAULT 'QUEUED'
        CHECK (status IN ('QUEUED', 'RUNNING', 'COMPLETED', 'FAILED', 'DEAD_LETTER')),
    attempt_count SMALLINT NOT NULL DEFAULT 0,
    error_message TEXT,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 32. post_media (Associates posts with media_assets for orphan GC tracking)
CREATE TABLE post_media (
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    media_asset_id UUID NOT NULL REFERENCES media_assets(id) ON DELETE RESTRICT,
    usage_role VARCHAR(24) NOT NULL DEFAULT 'BODY'
        CHECK (usage_role IN ('COVER', 'BODY', 'GALLERY', 'ATTACHMENT')),
    sort_order SMALLINT NOT NULL DEFAULT 0,
    PRIMARY KEY (post_id, media_asset_id, usage_role)
);

CREATE INDEX idx_media_assets_blog_hash ON media_assets(blog_id, sha256_hash);
CREATE INDEX idx_media_assets_orphan ON media_assets(orphan_since) WHERE reference_count = 0 AND status = 'READY';
