package community

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Fault is the module error envelope.
type Fault struct {
	Status  int
	Code    string
	Message string
}

func (f *Fault) Error() string { return f.Message }

// Service implements follows, subscriptions, reactions, comments,
// bookmarks, and user blocks.
type Service struct {
	pool *pgxpool.Pool
}

// NewService builds a community service.
func NewService(pool *pgxpool.Pool) (*Service, error) {
	if pool == nil {
		return nil, errors.New("community service requires a database pool")
	}
	return &Service{pool: pool}, nil
}

func (s *Service) requirePool() error {
	if s.pool == nil {
		return &Fault{Status: http.StatusServiceUnavailable, Code: "SERVICE_UNAVAILABLE", Message: "Database is unavailable"}
	}
	return nil
}

func (s *Service) blogExists(ctx context.Context, blogID string) error {
	var exists bool
	if err := s.pool.QueryRow(ctx, `SELECT TRUE FROM blogs WHERE id = $1`, blogID).Scan(&exists); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &Fault{Status: http.StatusNotFound, Code: "BLOG_NOT_FOUND", Message: "Blog not found"}
		}
		return err
	}
	return nil
}

// FollowedBlog is one blog the user follows.
type FollowedBlog struct {
	BlogID   string `json:"blog_id"`
	Handle   string `json:"handle"`
	Title    string `json:"title"`
	Relation string `json:"relation_type"`
}

// Follow creates a FOLLOW relation to a blog (idempotent).
func (s *Service) Follow(ctx context.Context, userID, blogID string) error {
	if err := s.requirePool(); err != nil {
		return err
	}
	if err := s.blogExists(ctx, blogID); err != nil {
		return err
	}
	_, err := s.pool.Exec(ctx, `INSERT INTO follows (follower_user_id, target_blog_id, relation_type, status)
		VALUES ($1, $2, 'FOLLOW', 'ACCEPTED')
		ON CONFLICT (follower_user_id, target_blog_id) DO UPDATE
		SET status = 'ACCEPTED', updated_at = NOW()`, userID, blogID)
	return err
}

// Unfollow terminates the follow relation (idempotent).
func (s *Service) Unfollow(ctx context.Context, userID, blogID string) error {
	if err := s.requirePool(); err != nil {
		return err
	}
	_, err := s.pool.Exec(ctx, `DELETE FROM follows WHERE follower_user_id = $1 AND target_blog_id = $2`, userID, blogID)
	return err
}

// ListFollows returns the blogs a user follows.
func (s *Service) ListFollows(ctx context.Context, userID string, limit int) ([]FollowedBlog, error) {
	if err := s.requirePool(); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	rows, err := s.pool.Query(ctx, `
		SELECT f.target_blog_id, b.handle, b.title, f.relation_type
		FROM follows f JOIN blogs b ON b.id = f.target_blog_id
		WHERE f.follower_user_id = $1 AND f.status = 'ACCEPTED'
		ORDER BY f.created_at DESC LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]FollowedBlog, 0)
	for rows.Next() {
		var item FollowedBlog
		if err := rows.Scan(&item.BlogID, &item.Handle, &item.Title, &item.Relation); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// Subscription is one blog subscription row.
type Subscription struct {
	BlogID   string `json:"blog_id"`
	Handle   string `json:"handle"`
	Title    string `json:"title"`
	Tier     string `json:"tier"`
	NotifyInApp bool `json:"notify_in_app"`
}

// Subscribe creates a FREE subscription (idempotent).
func (s *Service) Subscribe(ctx context.Context, userID, blogID string) error {
	if err := s.requirePool(); err != nil {
		return err
	}
	if err := s.blogExists(ctx, blogID); err != nil {
		return err
	}
	_, err := s.pool.Exec(ctx, `INSERT INTO subscriptions (subscriber_user_id, target_blog_id, tier, status)
		VALUES ($1, $2, 'FREE', 'ACTIVE')
		ON CONFLICT (subscriber_user_id, target_blog_id) DO UPDATE
		SET status = 'ACTIVE', updated_at = NOW()`, userID, blogID)
	return err
}

// Unsubscribe cancels the subscription (keeps history).
func (s *Service) Unsubscribe(ctx context.Context, userID, blogID string) error {
	if err := s.requirePool(); err != nil {
		return err
	}
	_, err := s.pool.Exec(ctx, `UPDATE subscriptions SET status = 'CANCELED', updated_at = NOW()
		WHERE subscriber_user_id = $1 AND target_blog_id = $2`, userID, blogID)
	return err
}

// ListSubscriptions returns active subscriptions.
func (s *Service) ListSubscriptions(ctx context.Context, userID string, limit int) ([]Subscription, error) {
	if err := s.requirePool(); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	rows, err := s.pool.Query(ctx, `
		SELECT s.target_blog_id, b.handle, b.title, s.tier, s.notify_in_app
		FROM subscriptions s JOIN blogs b ON b.id = s.target_blog_id
		WHERE s.subscriber_user_id = $1 AND s.status = 'ACTIVE'
		ORDER BY s.created_at DESC LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Subscription, 0)
	for rows.Next() {
		var item Subscription
		if err := rows.Scan(&item.BlogID, &item.Handle, &item.Title, &item.Tier, &item.NotifyInApp); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

var validReactionTypes = map[string]bool{
	"LIKE": true, "LOVE": true, "HELPFUL": true, "FUNNY": true, "TOUCHING": true, "INSIGHTFUL": true,
}

// ReactionSummary aggregates reaction counts for a post.
type ReactionSummary struct {
	Counts map[string]int `json:"counts"`
	Total  int            `json:"total"`
	Mine   *string        `json:"mine,omitempty"`
}

// React upserts the user's reaction to a post and maintains the counter.
func (s *Service) React(ctx context.Context, userID, postID, reactionType string) error {
	reactionType = strings.ToUpper(strings.TrimSpace(reactionType))
	if !validReactionTypes[reactionType] {
		return &Fault{Status: http.StatusBadRequest, Code: "INVALID_REACTION_TYPE", Message: "Unknown reaction type"}
	}
	if err := s.requirePool(); err != nil {
		return err
	}
	tx, err := s.DB().Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var allow bool
	err = tx.QueryRow(ctx, `SELECT allow_reactions FROM posts
		WHERE id = $1 AND status = 'PUBLISHED' AND trashed_at IS NULL`, postID).Scan(&allow)
	if errors.Is(err, pgx.ErrNoRows) {
		return &Fault{Status: http.StatusNotFound, Code: "POST_NOT_FOUND", Message: "Post not found"}
	}
	if err != nil {
		return err
	}
	if !allow {
		return &Fault{Status: http.StatusForbidden, Code: "REACTIONS_DISABLED", Message: "Reactions are disabled for this post"}
	}
	// xmax = 0 is true only for freshly inserted rows, distinguishing
	// first-time reactions (counter +1) from reaction-type changes.
	var inserted bool
	err = tx.QueryRow(ctx, `INSERT INTO reactions (target_type, target_id, user_id, reaction_type)
		VALUES ('POST', $1, $2, $3)
		ON CONFLICT (target_type, target_id, user_id) DO UPDATE SET reaction_type = EXCLUDED.reaction_type
		RETURNING (xmax = 0)`, postID, userID, reactionType).Scan(&inserted)
	if err != nil {
		return err
	}
	if inserted {
		if _, err := tx.Exec(ctx, `UPDATE posts SET reaction_count = reaction_count + 1 WHERE id = $1`, postID); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// Unreact removes the user's reaction and maintains the counter.
func (s *Service) Unreact(ctx context.Context, userID, postID string) error {
	if err := s.requirePool(); err != nil {
		return err
	}
	tx, err := s.DB().Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	res, err := tx.Exec(ctx, `DELETE FROM reactions
		WHERE target_type = 'POST' AND target_id = $1 AND user_id = $2`, postID, userID)
	if err != nil {
		return err
	}
	if res.RowsAffected() > 0 {
		_, _ = tx.Exec(ctx, `UPDATE posts SET reaction_count = GREATEST(reaction_count - 1, 0) WHERE id = $1`, postID)
	}
	return tx.Commit(ctx)
}

// ReactionCounts aggregates counts for a post.
func (s *Service) ReactionCounts(ctx context.Context, userID, postID string) (*ReactionSummary, error) {
	if err := s.requirePool(); err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, `SELECT reaction_type, COUNT(*)
		FROM reactions WHERE target_type = 'POST' AND target_id = $1 GROUP BY reaction_type`, postID)
	if err != nil {
		return nil, err
	}
	summary := &ReactionSummary{Counts: map[string]int{}}
	for rows.Next() {
		var kind string
		var count int
		if err := rows.Scan(&kind, &count); err != nil {
			rows.Close()
			return nil, err
		}
		summary.Counts[kind] = count
		summary.Total += count
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if userID != "" {
		var mine string
		if err := s.pool.QueryRow(ctx, `SELECT reaction_type FROM reactions
			WHERE target_type = 'POST' AND target_id = $1 AND user_id = $2`,
			postID, userID).Scan(&mine); err == nil {
			summary.Mine = &mine
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
	}
	return summary, nil
}

// DB exposes the pool for transactional helpers.
func (s *Service) DB() *pgxpool.Pool {
	return s.pool
}
