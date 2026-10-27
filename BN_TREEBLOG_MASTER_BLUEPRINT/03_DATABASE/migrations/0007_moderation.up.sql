-- Migration: 0007_moderation.up.sql
-- Domain: Trust, Safety & Moderation (5 Tables: user_reports, moderation_cases, moderation_actions, appeals, spam_detection_results)

-- 47. moderation_cases
CREATE TABLE moderation_cases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    target_type VARCHAR(24) NOT NULL
        CHECK (target_type IN ('POST', 'COMMENT', 'GUESTBOOK', 'BLOG', 'USER', 'AD_CREATIVE')),
    target_id UUID NOT NULL,
    reason_category VARCHAR(32) NOT NULL
        CHECK (reason_category IN ('SPAM', 'COPYRIGHT_DMCA', 'DEFAMATION', 'PRIVACY_LEAK', 'ADULT_ILLEGAL', 'PHISHING_MALWARE', 'IMPERSONATION', 'HATE_SPEECH')),
    report_count INTEGER NOT NULL DEFAULT 1 CHECK (report_count >= 1),
    priority_score SMALLINT NOT NULL DEFAULT 50 CHECK (priority_score BETWEEN 1 AND 100),
    status VARCHAR(24) NOT NULL DEFAULT 'OPEN'
        CHECK (status IN ('OPEN', 'IN_REVIEW', 'ACTION_TAKEN', 'DISMISSED', 'APPEALED', 'APPEAL_RESOLVED')),
    assigned_moderator_id UUID REFERENCES users(id) ON DELETE SET NULL,
    snapshot_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (target_type, target_id, reason_category)
);

-- 48. user_reports
CREATE TABLE user_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    case_id UUID NOT NULL REFERENCES moderation_cases(id) ON DELETE CASCADE,
    reporter_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    reason_category VARCHAR(32) NOT NULL,
    description VARCHAR(2000) NOT NULL DEFAULT '',
    evidence_urls_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    is_copyright_owner_confirmed BOOLEAN NOT NULL DEFAULT FALSE,
    reporter_ip INET,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 49. moderation_actions
CREATE TABLE moderation_actions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    case_id UUID NOT NULL REFERENCES moderation_cases(id) ON DELETE RESTRICT,
    actor_moderator_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    action_type VARCHAR(32) NOT NULL
        CHECK (action_type IN ('DISMISS', 'WARN_USER', 'QUARANTINE_CONTENT', 'RESTORE_CONTENT', 'DELETE_CONTENT', 'SUSPEND_BLOG_7D', 'SUSPEND_BLOG_PERM', 'SUSPEND_USER')),
    policy_clause_code VARCHAR(40) NOT NULL,
    internal_note TEXT NOT NULL DEFAULT '',
    public_notice_reason VARCHAR(500) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 50. appeals
CREATE TABLE appeals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    case_id UUID NOT NULL REFERENCES moderation_cases(id) ON DELETE RESTRICT,
    appellant_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    appeal_statement TEXT NOT NULL,
    supporting_docs_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    status VARCHAR(24) NOT NULL DEFAULT 'PENDING'
        CHECK (status IN ('PENDING', 'UNDER_REVIEW', 'APPROVED_RESTORED', 'REJECTED_UPHELD')),
    reviewer_moderator_id UUID REFERENCES users(id) ON DELETE SET NULL,
    reviewer_decision_note VARCHAR(1000),
    resolved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 51. spam_detection_results
CREATE TABLE spam_detection_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    target_type VARCHAR(24) NOT NULL CHECK (target_type IN ('POST', 'COMMENT', 'BLOG')),
    target_id UUID NOT NULL,
    detector_version VARCHAR(32) NOT NULL,
    spam_score NUMERIC(4,3) NOT NULL CHECK (spam_score BETWEEN 0 AND 1),
    matched_rules_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    auto_action_taken VARCHAR(24) NOT NULL DEFAULT 'NONE'
        CHECK (auto_action_taken IN ('NONE', 'FLAGGED_FOR_REVIEW', 'AUTO_QUARANTINED')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_mod_cases_queue ON moderation_cases(status, priority_score DESC, created_at ASC);
