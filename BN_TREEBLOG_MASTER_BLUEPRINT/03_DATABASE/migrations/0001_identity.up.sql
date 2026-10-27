-- Migration: 0001_identity.up.sql
-- Domain: Identity & Authorization (9 Tables: users, user_profiles, identities, sessions, roles, permissions, role_permissions, user_roles, user_consents)

CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "citext";

-- 1. users
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email CITEXT NOT NULL UNIQUE,
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    password_hash TEXT,
    status VARCHAR(24) NOT NULL DEFAULT 'ACTIVE'
        CHECK (status IN ('PENDING_VERIFICATION', 'ACTIVE', 'SUSPENDED', 'DEACTIVATING', 'DELETED')),
    mfa_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    mfa_secret_enc BYTEA,
    mfa_backup_codes_hash JSONB NOT NULL DEFAULT '[]'::jsonb,
    preferred_locale VARCHAR(16) NOT NULL DEFAULT 'ko-KR',
    timezone VARCHAR(64) NOT NULL DEFAULT 'Asia/Seoul',
    trust_tier SMALLINT NOT NULL DEFAULT 1 CHECK (trust_tier BETWEEN 0 AND 5),
    last_login_at TIMESTAMPTZ,
    deactivation_requested_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 2. user_profiles
CREATE TABLE user_profiles (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    username CITEXT NOT NULL UNIQUE CHECK (char_length(username) BETWEEN 3 AND 32),
    display_name VARCHAR(80) NOT NULL,
    bio VARCHAR(500) NOT NULL DEFAULT '',
    avatar_media_id UUID,
    website_url VARCHAR(512),
    social_links_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    is_public_profile BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 3. identities (OAuth / OIDC / Local / TreeScan Account Linking)
CREATE TABLE identities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(32) NOT NULL
        CHECK (provider IN ('LOCAL', 'GOOGLE', 'APPLE', 'GITHUB', 'NAVER', 'KAKAO', 'TREESCAN_OIDC')),
    provider_subject VARCHAR(255) NOT NULL,
    provider_email CITEXT,
    metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    linked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMPTZ,
    UNIQUE (provider, provider_subject)
);

-- 4. sessions
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash CHAR(64) NOT NULL UNIQUE,
    token_family_id UUID NOT NULL,
    user_agent VARCHAR(512) NOT NULL DEFAULT '',
    ip_address INET,
    device_name VARCHAR(120),
    mfa_verified BOOLEAN NOT NULL DEFAULT FALSE,
    is_revoked BOOLEAN NOT NULL DEFAULT FALSE,
    revoked_reason VARCHAR(64),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_active_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 5. roles
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(40) NOT NULL UNIQUE
        CHECK (code IN ('USER', 'ADVERTISER', 'SUPPORT_AGENT', 'TRUST_SAFETY_MOD', 'AD_REVIEWER', 'FINANCE_ADMIN', 'SUPER_ADMIN')),
    name VARCHAR(80) NOT NULL,
    description VARCHAR(300) NOT NULL DEFAULT '',
    is_system BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 6. permissions
CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(80) NOT NULL UNIQUE,
    module VARCHAR(40) NOT NULL,
    description VARCHAR(300) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 7. role_permissions
CREATE TABLE role_permissions (
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    granted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (role_id, permission_id)
);

-- 8. user_roles
CREATE TABLE user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    granted_by UUID REFERENCES users(id) ON DELETE SET NULL,
    granted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id)
);

-- 9. user_consents (Terms, Privacy, Marketing, Content Import Ownership, TreeScan Data Sync)
CREATE TABLE user_consents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    consent_type VARCHAR(48) NOT NULL
        CHECK (consent_type IN ('TERMS_OF_SERVICE', 'PRIVACY_POLICY', 'MARKETING_EMAIL', 'MARKETING_PUSH', 'CONTENT_IMPORT_OWNERSHIP', 'TREESCAN_PROFILE_SYNC')),
    policy_version VARCHAR(24) NOT NULL,
    is_agreed BOOLEAN NOT NULL,
    ip_address INET,
    user_agent VARCHAR(512),
    agreed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_identities_user_id ON identities(user_id);
CREATE INDEX idx_sessions_user_id_active ON sessions(user_id, expires_at) WHERE is_revoked = FALSE;
CREATE INDEX idx_user_consents_user_type ON user_consents(user_id, consent_type);
