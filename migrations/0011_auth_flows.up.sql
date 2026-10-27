-- Credential recovery and abuse controls required by TB-AUTH-001.
CREATE TABLE email_verification_challenges (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    code_hash CHAR(64) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    attempts SMALLINT NOT NULL DEFAULT 0,
    last_sent_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE password_reset_challenges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash CHAR(64) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_password_reset_user ON password_reset_challenges(user_id, created_at DESC);

CREATE TABLE login_attempts (
    email CITEXT PRIMARY KEY,
    failed_count SMALLINT NOT NULL DEFAULT 0,
    locked_until TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
