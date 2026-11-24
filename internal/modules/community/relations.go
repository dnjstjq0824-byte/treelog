package community

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

const (
	// maxActiveFollowsPerUser caps ACCEPTED + PENDING relations per user as
	// abuse friction. No blueprint clause sets this number; raise it when
	// legitimate power users hit the ceiling.
	maxActiveFollowsPerUser = 1000
	// maxGroupsPerUser caps neighbor folders per user.
	maxGroupsPerUser = 20
)

// NeighborRequest is one pending neighbor application with applicant info.
type NeighborRequest struct {
	FollowID    string    `json:"follow_id"`
	UserID      string    `json:"user_id"`
	DisplayName string    `json:"display_name"`
	Message     string    `json:"message"`
	CreatedAt   time.Time `json:"created_at"`
}

// NeighborGroup is one user-owned neighbor folder.
type NeighborGroup struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
}

// FeedItem is one 모아보기 (following feed) entry.
type FeedItem struct {
	PostID      string     `json:"post_id"`
	BlogID      string     `json:"blog_id"`
	BlogHandle  string     `json:"blog_handle"`
	BlogTitle   string     `json:"blog_title"`
	Slug        string     `json:"slug"`
	Title       string     `json:"title"`
	Excerpt     string     `json:"excerpt"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
}

// RequestNeighbor files a NEIGHBOR application to a blog. Re-requesting
// after a rejection or termination reopens the request.
func (s *Service) RequestNeighbor(ctx context.Context, userID, blogID, message string) error {
	if err := s.requirePool(); err != nil {
		return err
	}
	if len(message) > 300 {
		return &Fault{Status: http.StatusBadRequest, Code: "INVALID_REQUEST_MESSAGE", Message: "Request message exceeds 300 characters"}
	}
	var ownerID string
	if err := s.pool.QueryRow(ctx, `SELECT owner_user_id FROM blogs WHERE id = $1`, blogID).Scan(&ownerID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &Fault{Status: http.StatusNotFound, Code: "BLOG_NOT_FOUND", Message: "Blog not found"}
		}
		return err
	}
	if ownerID == userID {
		return &Fault{Status: http.StatusBadRequest, Code: "CANNOT_NEIGHBOR_OWN_BLOG", Message: "You cannot neighbor your own blog"}
	}
	var active int
	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM follows
		WHERE follower_user_id = $1 AND status IN ('PENDING', 'ACCEPTED')`, userID).Scan(&active); err != nil {
		return err
	}
	if active >= maxActiveFollowsPerUser {
		return &Fault{Status: http.StatusForbidden, Code: "FOLLOW_LIMIT_REACHED", Message: "Neighbor limit reached"}
	}
	res, err := s.pool.Exec(ctx, `
		INSERT INTO follows (follower_user_id, target_blog_id, relation_type, status, request_message)
		VALUES ($1, $2, 'NEIGHBOR', 'PENDING', $3)
		ON CONFLICT (follower_user_id, target_blog_id) DO UPDATE
		SET relation_type = 'NEIGHBOR', status = 'PENDING', request_message = EXCLUDED.request_message,
		    responded_at = NULL, updated_at = NOW()
		WHERE follows.status IN ('REJECTED', 'TERMINATED')
		OR (follows.status = 'PENDING' AND follows.relation_type = 'NEIGHBOR')`, userID, blogID, strings.TrimSpace(message))
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return &Fault{Status: http.StatusConflict, Code: "ALREADY_NEIGHBOR", Message: "A neighbor relation already exists"}
	}
	s.notifyBlogFollowed(ctx, blogID, userID, s.displayName(ctx, userID), "NEIGHBOR", "이웃 신청")
	return nil
}

// ListNeighborRequests returns pending applications oldest-first for moderators.
func (s *Service) ListNeighborRequests(ctx context.Context, userID, blogID string) ([]NeighborRequest, error) {
	if err := s.requirePool(); err != nil {
		return nil, err
	}
	if !s.canModerate(ctx, userID, blogID) {
		return nil, &Fault{Status: http.StatusForbidden, Code: "NOT_BLOG_MODERATOR", Message: "Only blog moderators can review neighbor requests"}
	}
	rows, err := s.pool.Query(ctx, `
		SELECT f.id, f.follower_user_id, COALESCE(up.display_name, ''), f.request_message, f.created_at
		FROM follows f
		LEFT JOIN user_profiles up ON up.user_id = f.follower_user_id
		WHERE f.target_blog_id = $1 AND f.relation_type = 'NEIGHBOR' AND f.status = 'PENDING'
		ORDER BY f.created_at ASC`, blogID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]NeighborRequest, 0)
	for rows.Next() {
		var item NeighborRequest
		if err := rows.Scan(&item.FollowID, &item.UserID, &item.DisplayName, &item.Message, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// RespondNeighborRequest accepts or rejects a pending application. Accepting
// promotes both sides to MUTUAL_NEIGHBOR when the blog owner already
// neighbors any blog owned by the applicant.
func (s *Service) RespondNeighborRequest(ctx context.Context, userID, followID string, accept bool) error {
	if err := s.requirePool(); err != nil {
		return err
	}
	tx, err := s.DB().Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var blogID, applicantID, relation, status string
	if err := tx.QueryRow(ctx, `SELECT target_blog_id, follower_user_id, relation_type, status
		FROM follows WHERE id = $1 FOR UPDATE`, followID).Scan(&blogID, &applicantID, &relation, &status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &Fault{Status: http.StatusNotFound, Code: "REQUEST_NOT_FOUND", Message: "Neighbor request not found"}
		}
		return err
	}
	if !s.canModerateTx(ctx, tx, userID, blogID) {
		return &Fault{Status: http.StatusForbidden, Code: "NOT_BLOG_MODERATOR", Message: "Only blog moderators can review neighbor requests"}
	}
	if relation != "NEIGHBOR" || status != "PENDING" {
		return &Fault{Status: http.StatusConflict, Code: "REQUEST_NOT_PENDING", Message: "Request is not pending"}
	}
	next := "REJECTED"
	if accept {
		next = "ACCEPTED"
	}
	if _, err := tx.Exec(ctx, `UPDATE follows SET status = $1, responded_at = NOW(), updated_at = NOW()
		WHERE id = $2`, next, followID); err != nil {
		return err
	}
	if accept {
		// Mutual upgrade: applicant's blogs already neighbored by this owner?
		var reverseID string
		err := tx.QueryRow(ctx, `
			SELECT f.id FROM follows f
			JOIN blogs b ON b.id = f.target_blog_id
			WHERE f.follower_user_id = $1 AND b.owner_user_id = $2
			  AND f.relation_type = 'NEIGHBOR' AND f.status = 'ACCEPTED'
			LIMIT 1`, userID, applicantID).Scan(&reverseID)
		if err == nil {
			_, _ = tx.Exec(ctx, `UPDATE follows SET relation_type = 'MUTUAL_NEIGHBOR', updated_at = NOW()
				WHERE id IN ($1, $2)`, followID, reverseID)
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
	}
	return tx.Commit(ctx)
}

// CreateNeighborGroup creates a neighbor folder for the user.
func (s *Service) CreateNeighborGroup(ctx context.Context, userID, name string) (*NeighborGroup, error) {
	if err := s.requirePool(); err != nil {
		return nil, err
	}
	name = strings.TrimSpace(name)
	if name == "" || len([]rune(name)) > 60 {
		return nil, &Fault{Status: http.StatusBadRequest, Code: "INVALID_GROUP_NAME", Message: "Group name must be 1-60 characters"}
	}
	var count int
	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM neighbor_groups WHERE owner_user_id = $1`, userID).Scan(&count); err != nil {
		return nil, err
	}
	if count >= maxGroupsPerUser {
		return nil, &Fault{Status: http.StatusForbidden, Code: "GROUP_LIMIT_REACHED", Message: "Neighbor group limit reached"}
	}
	var group NeighborGroup
	if err := s.pool.QueryRow(ctx, `
		INSERT INTO neighbor_groups (owner_user_id, name, sort_order)
		SELECT $1, $2, COALESCE(MAX(sort_order), -1) + 1 FROM neighbor_groups WHERE owner_user_id = $1
		RETURNING id, name, sort_order, created_at
	`, userID, name).Scan(&group.ID, &group.Name, &group.SortOrder, &group.CreatedAt); err != nil {
		return nil, err
	}
	return &group, nil
}

// ListNeighborGroups returns the user's folders in sort order.
func (s *Service) ListNeighborGroups(ctx context.Context, userID string) ([]NeighborGroup, error) {
	if err := s.requirePool(); err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, `SELECT id, name, sort_order, created_at FROM neighbor_groups
		WHERE owner_user_id = $1 ORDER BY sort_order ASC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	groups := make([]NeighborGroup, 0)
	for rows.Next() {
		var group NeighborGroup
		if err := rows.Scan(&group.ID, &group.Name, &group.SortOrder, &group.CreatedAt); err != nil {
			return nil, err
		}
		groups = append(groups, group)
	}
	return groups, rows.Err()
}

// RenameNeighborGroup renames a folder owned by the user.
func (s *Service) RenameNeighborGroup(ctx context.Context, userID, groupID, name string) error {
	if err := s.requirePool(); err != nil {
		return err
	}
	name = strings.TrimSpace(name)
	if name == "" || len([]rune(name)) > 60 {
		return &Fault{Status: http.StatusBadRequest, Code: "INVALID_GROUP_NAME", Message: "Group name must be 1-60 characters"}
	}
	res, err := s.pool.Exec(ctx, `UPDATE neighbor_groups SET name = $1 WHERE id = $2 AND owner_user_id = $3`,
		name, groupID, userID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return &Fault{Status: http.StatusNotFound, Code: "GROUP_NOT_FOUND", Message: "Neighbor group not found"}
	}
	return nil
}

// DeleteNeighborGroup removes a folder; members fall back to ungrouped.
func (s *Service) DeleteNeighborGroup(ctx context.Context, userID, groupID string) error {
	if err := s.requirePool(); err != nil {
		return err
	}
	res, err := s.pool.Exec(ctx, `DELETE FROM neighbor_groups WHERE id = $1 AND owner_user_id = $2`, groupID, userID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return &Fault{Status: http.StatusNotFound, Code: "GROUP_NOT_FOUND", Message: "Neighbor group not found"}
	}
	return nil
}

// MoveFollowToGroup files the user's relation to a blog into a folder.
// A nil group clears the assignment.
func (s *Service) MoveFollowToGroup(ctx context.Context, userID, blogID string, groupID *string) error {
	if err := s.requirePool(); err != nil {
		return err
	}
	if groupID != nil && *groupID != "" {
		var owner string
		if err := s.pool.QueryRow(ctx, `SELECT owner_user_id FROM neighbor_groups WHERE id = $1`, *groupID).Scan(&owner); err != nil || owner != userID {
			if errors.Is(err, pgx.ErrNoRows) || owner != userID {
				return &Fault{Status: http.StatusNotFound, Code: "GROUP_NOT_FOUND", Message: "Neighbor group not found"}
			}
			return err
		}
	} else {
		groupID = nil
	}
	res, err := s.pool.Exec(ctx, `UPDATE follows SET neighbor_group_id = $1, updated_at = NOW()
		WHERE follower_user_id = $2 AND target_blog_id = $3 AND status = 'ACCEPTED'`, groupID, userID, blogID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return &Fault{Status: http.StatusNotFound, Code: "FOLLOW_NOT_FOUND", Message: "No accepted relation to this blog"}
	}
	return nil
}

// SetFollowFavorite toggles the favorite star on an accepted relation.
func (s *Service) SetFollowFavorite(ctx context.Context, userID, blogID string, favorite bool) error {
	if err := s.requirePool(); err != nil {
		return err
	}
	res, err := s.pool.Exec(ctx, `UPDATE follows SET is_favorite = $1, updated_at = NOW()
		WHERE follower_user_id = $2 AND target_blog_id = $3 AND status = 'ACCEPTED'`, favorite, userID, blogID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return &Fault{Status: http.StatusNotFound, Code: "FOLLOW_NOT_FOUND", Message: "No accepted relation to this blog"}
	}
	return nil
}

// PauseSubscription suspends delivery without canceling history.
func (s *Service) PauseSubscription(ctx context.Context, userID, blogID string) error {
	return s.setSubscriptionStatus(ctx, userID, blogID, "ACTIVE", "PAUSED")
}

// ResumeSubscription reactivates a paused subscription.
func (s *Service) ResumeSubscription(ctx context.Context, userID, blogID string) error {
	return s.setSubscriptionStatus(ctx, userID, blogID, "PAUSED", "ACTIVE")
}

func (s *Service) setSubscriptionStatus(ctx context.Context, userID, blogID, from, to string) error {
	if err := s.requirePool(); err != nil {
		return err
	}
	res, err := s.pool.Exec(ctx, `UPDATE subscriptions SET status = $1, updated_at = NOW()
		WHERE subscriber_user_id = $2 AND target_blog_id = $3 AND status = $4`, to, userID, blogID, from)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return &Fault{Status: http.StatusConflict, Code: "SUBSCRIPTION_STATE_MISMATCH", Message: "Subscription is not in the expected state"}
	}
	return nil
}

// UpdateSubscriptionChannels toggles per-channel delivery for a subscription.
// Nil fields keep their current value.
func (s *Service) UpdateSubscriptionChannels(ctx context.Context, userID, blogID string, inApp, webPush, email *bool) error {
	if err := s.requirePool(); err != nil {
		return err
	}
	res, err := s.pool.Exec(ctx, `UPDATE subscriptions
		SET notify_in_app = COALESCE($1, notify_in_app),
		    notify_web_push = COALESCE($2, notify_web_push),
		    notify_email = COALESCE($3, notify_email),
		    updated_at = NOW()
		WHERE subscriber_user_id = $4 AND target_blog_id = $5`, inApp, webPush, email, userID, blogID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return &Fault{Status: http.StatusNotFound, Code: "SUBSCRIPTION_NOT_FOUND", Message: "Subscription not found"}
	}
	return nil
}

// FollowingFeed returns newest-first public posts from followed and
// subscribed blogs. Cursor format: RFC3339Nano|postID (same shape as the
// notification cursor).
func (s *Service) FollowingFeed(ctx context.Context, userID string, limit int, cursor string) ([]FeedItem, string, error) {
	if err := s.requirePool(); err != nil {
		return nil, "", err
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	query := `
		SELECT p.id, p.blog_id, b.handle, b.title, p.slug, p.title, p.excerpt, p.published_at
		FROM posts p JOIN blogs b ON b.id = p.blog_id
		WHERE p.status = 'PUBLISHED' AND p.visibility = 'PUBLIC' AND p.trashed_at IS NULL
		  AND b.status = 'ACTIVE'
		  AND (EXISTS (SELECT 1 FROM follows f
		         WHERE f.follower_user_id = $1 AND f.target_blog_id = p.blog_id AND f.status = 'ACCEPTED')
		    OR EXISTS (SELECT 1 FROM subscriptions s
		         WHERE s.subscriber_user_id = $1 AND s.target_blog_id = p.blog_id AND s.status = 'ACTIVE'))
	`
	args := []any{userID}
	if cursor != "" {
		moment, id, err := decodeFeedCursor(cursor)
		if err != nil {
			return nil, "", &Fault{Status: http.StatusBadRequest, Code: "INVALID_CURSOR", Message: "Invalid pagination cursor"}
		}
		query += ` AND (COALESCE(p.published_at, p.created_at), p.id) < ($2, $3)`
		args = append(args, moment, id)
	}
	query += ` ORDER BY COALESCE(p.published_at, p.created_at) DESC, p.id DESC LIMIT $` + strconv.Itoa(len(args)+1)
	args = append(args, limit+1)
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()
	items := make([]FeedItem, 0, limit)
	for rows.Next() {
		var item FeedItem
		if err := rows.Scan(&item.PostID, &item.BlogID, &item.BlogHandle, &item.BlogTitle,
			&item.Slug, &item.Title, &item.Excerpt, &item.PublishedAt); err != nil {
			return nil, "", err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, "", err
	}
	nextCursor := ""
	if len(items) > limit {
		last := items[limit-1]
		items = items[:limit]
		moment := time.Now().UTC()
		if last.PublishedAt != nil {
			moment = *last.PublishedAt
		}
		nextCursor = encodeFeedCursor(moment, last.PostID)
	}
	return items, nextCursor, nil
}

func encodeFeedCursor(moment time.Time, id string) string {
	raw := moment.UTC().Format(time.RFC3339Nano) + "|" + id
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func decodeFeedCursor(cursor string) (time.Time, string, error) {
	raw, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return time.Time{}, "", err
	}
	parts := strings.SplitN(string(raw), "|", 2)
	if len(parts) != 2 || parts[1] == "" {
		return time.Time{}, "", errors.New("invalid cursor shape")
	}
	moment, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return time.Time{}, "", err
	}
	return moment, parts[1], nil
}
