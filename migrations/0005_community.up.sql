-- Migration: 0005_community.up.sql
-- Domain: Engagement, Community & Notifications (10 Tables: neighbor_groups, follows, subscriptions, reactions, comments, bookmarks, guestbook_entries, user_blocks, notification_preferences, notifications)

-- 33. neighbor_groups
CREATE TABLE neighbor_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(60) NOT NULL,
    sort_order SMALLINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (owner_user_id, name)
);

-- 34. follows (Follow / Neighbor / Mutual Neighbor State Machine)
CREATE TABLE follows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    follower_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_blog_id UUID NOT NULL REFERENCES blogs(id) ON DELETE CASCADE,
    neighbor_group_id UUID REFERENCES neighbor_groups(id) ON DELETE SET NULL,
    relation_type VARCHAR(24) NOT NULL DEFAULT 'FOLLOW'
        CHECK (relation_type IN ('FOLLOW', 'NEIGHBOR', 'MUTUAL_NEIGHBOR')),
    status VARCHAR(24) NOT NULL DEFAULT 'ACCEPTED'
        CHECK (status IN ('PENDING', 'ACCEPTED', 'REJECTED', 'TERMINATED')),
    request_message VARCHAR(300) NOT NULL DEFAULT '',
    is_favorite BOOLEAN NOT NULL DEFAULT FALSE,
    responded_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (follower_user_id, target_blog_id)
);

-- 35. subscriptions (Notification & Newsletter Delivery Contract)
CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subscriber_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_blog_id UUID NOT NULL REFERENCES blogs(id) ON DELETE CASCADE,
    tier VARCHAR(24) NOT NULL DEFAULT 'FREE'
        CHECK (tier IN ('FREE', 'PAID_TIER')),
    notify_in_app BOOLEAN NOT NULL DEFAULT TRUE,
    notify_web_push BOOLEAN NOT NULL DEFAULT TRUE,
    notify_email BOOLEAN NOT NULL DEFAULT FALSE,
    status VARCHAR(24) NOT NULL DEFAULT 'ACTIVE'
        CHECK (status IN ('ACTIVE', 'PAUSED', 'CANCELED', 'EXPIRED')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (subscriber_user_id, target_blog_id)
);

-- 36. reactions (6 Reaction Types on Posts and LIKE on Comments)
CREATE TABLE reactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    target_type VARCHAR(16) NOT NULL CHECK (target_type IN ('POST', 'COMMENT')),
    target_id UUID NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reaction_type VARCHAR(24) NOT NULL DEFAULT 'LIKE'
        CHECK (reaction_type IN ('LIKE', 'LOVE', 'HELPFUL', 'FUNNY', 'TOUCHING', 'INSIGHTFUL')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (target_type, target_id, user_id)
);

-- 37. comments (2-level Nested Comments with Secret & Approval Support)
CREATE TABLE comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    blog_id UUID NOT NULL REFERENCES blogs(id) ON DELETE CASCADE,
    author_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    parent_comment_id UUID REFERENCES comments(id) ON DELETE CASCADE,
    depth SMALLINT NOT NULL DEFAULT 0 CHECK (depth IN (0, 1)),
    content_text VARCHAR(3000) NOT NULL,
    is_secret BOOLEAN NOT NULL DEFAULT FALSE,
    is_pinned BOOLEAN NOT NULL DEFAULT FALSE,
    status VARCHAR(24) NOT NULL DEFAULT 'PUBLISHED'
        CHECK (status IN ('PENDING_APPROVAL', 'PUBLISHED', 'HIDDEN_BY_OWNER', 'DELETED_BY_USER', 'QUARANTINED_SPAM')),
    like_count INTEGER NOT NULL DEFAULT 0 CHECK (like_count >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 38. bookmarks
CREATE TABLE bookmarks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    folder_name VARCHAR(80) NOT NULL DEFAULT '기본 보관함',
    read_progress_pct SMALLINT NOT NULL DEFAULT 0 CHECK (read_progress_pct BETWEEN 0 AND 100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, post_id)
);

-- 39. guestbook_entries
CREATE TABLE guestbook_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    blog_id UUID NOT NULL REFERENCES blogs(id) ON DELETE CASCADE,
    author_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    parent_entry_id UUID REFERENCES guestbook_entries(id) ON DELETE CASCADE,
    content_text VARCHAR(2000) NOT NULL,
    is_secret BOOLEAN NOT NULL DEFAULT FALSE,
    status VARCHAR(24) NOT NULL DEFAULT 'PUBLISHED'
        CHECK (status IN ('PUBLISHED', 'DELETED_BY_USER', 'DELETED_BY_OWNER', 'QUARANTINED')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 40. user_blocks
CREATE TABLE user_blocks (
    blocker_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    blocked_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE CHECK (blocker_user_id <> blocked_user_id),
    reason VARCHAR(240) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (blocker_user_id, blocked_user_id)
);

-- 41. notification_preferences
CREATE TABLE notification_preferences (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    comments_in_app BOOLEAN NOT NULL DEFAULT TRUE,
    comments_push BOOLEAN NOT NULL DEFAULT TRUE,
    comments_email BOOLEAN NOT NULL DEFAULT FALSE,
    reactions_in_app BOOLEAN NOT NULL DEFAULT TRUE,
    reactions_push BOOLEAN NOT NULL DEFAULT FALSE,
    neighbors_in_app BOOLEAN NOT NULL DEFAULT TRUE,
    neighbors_push BOOLEAN NOT NULL DEFAULT TRUE,
    neighbors_email BOOLEAN NOT NULL DEFAULT TRUE,
    subscriptions_in_app BOOLEAN NOT NULL DEFAULT TRUE,
    subscriptions_push BOOLEAN NOT NULL DEFAULT TRUE,
    subscriptions_email BOOLEAN NOT NULL DEFAULT TRUE,
    quiet_hours_start TIME,
    quiet_hours_end TIME,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 42. notifications (Partitioned by created_at monthly in production)
CREATE TABLE notifications (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    recipient_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category VARCHAR(24) NOT NULL
        CHECK (category IN ('COMMENT', 'REACTION', 'NEIGHBOR', 'SUBSCRIPTION', 'MODERATION', 'MONETIZATION', 'SYSTEM', 'SECURITY')),
    group_key VARCHAR(128) NOT NULL,
    dedup_key VARCHAR(160) NOT NULL,
    actor_count INTEGER NOT NULL DEFAULT 1 CHECK (actor_count >= 1),
    actors_preview_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    title VARCHAR(240) NOT NULL,
    body VARCHAR(500) NOT NULL,
    action_url VARCHAR(512) NOT NULL,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

CREATE TABLE notifications_2026_10 PARTITION OF notifications
    FOR VALUES FROM ('2026-10-01 00:00:00+00') TO ('2026-11-01 00:00:00+00');
CREATE TABLE notifications_default PARTITION OF notifications DEFAULT;

CREATE INDEX idx_follows_target_blog ON follows(target_blog_id, relation_type, status);
CREATE INDEX idx_comments_post_created ON comments(post_id, created_at ASC);
CREATE INDEX idx_notifications_recipient_unread ON notifications(recipient_user_id, created_at DESC) WHERE is_read = FALSE;
