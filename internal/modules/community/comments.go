package community

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// Comment is one comment row with its author display name.
type Comment struct {
	ID           string    `json:"id"`
	PostID       string    `json:"post_id"`
	AuthorUserID string    `json:"author_user_id"`
	AuthorName   string    `json:"author_name"`
	ParentID     *string   `json:"parent_comment_id,omitempty"`
	Depth        int       `json:"depth"`
	ContentText  string    `json:"content_text"`
	IsSecret     bool      `json:"is_secret"`
	IsPinned     bool      `json:"is_pinned"`
	LikeCount    int       `json:"like_count"`
	CreatedAt    time.Time `json:"created_at"`
}

// CreateCommentInput carries the comment creation payload.
type CreateCommentInput struct {
	ContentText     string  `json:"content_text"`
	ParentCommentID *string `json:"parent_comment_id,omitempty"`
	IsSecret        bool    `json:"is_secret"`
}

// CreateComment writes a comment honoring the post's allow_comments policy.
func (s *Service) CreateComment(ctx context.Context, userID, postID string, input CreateCommentInput) (*Comment, error) {
	content := strings.TrimSpace(input.ContentText)
	if content == "" || len(content) > 3000 {
		return nil, &Fault{Status: http.StatusBadRequest, Code: "INVALID_COMMENT", Message: "Comment must be 1-3000 characters"}
	}
	if err := s.requirePool(); err != nil {
		return nil, err
	}
	tx, err := s.DB().Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var blogID, allowComments, status, visibility string
	err = tx.QueryRow(ctx, `SELECT blog_id, allow_comments, status, visibility FROM posts
		WHERE id = $1 AND trashed_at IS NULL`, postID).Scan(&blogID, &allowComments, &status, &visibility)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, &Fault{Status: http.StatusNotFound, Code: "POST_NOT_FOUND", Message: "Post not found"}
	}
	if err != nil {
		return nil, err
	}
	if status != "PUBLISHED" {
		return nil, &Fault{Status: http.StatusForbidden, Code: "POST_NOT_PUBLISHED", Message: "Comments are closed for unpublished posts"}
	}
	if visibility == "PRIVATE" {
		var authorID string
		_ = tx.QueryRow(ctx, `SELECT author_user_id FROM posts WHERE id = $1`, postID).Scan(&authorID)
		if authorID != userID {
			return nil, &Fault{Status: http.StatusForbidden, Code: "POST_PRIVATE", Message: "This post is private"}
		}
	}
	commentStatus := "PUBLISHED"
	switch allowComments {
	case "DISABLED":
		return nil, &Fault{Status: http.StatusForbidden, Code: "COMMENTS_DISABLED", Message: "Comments are disabled for this post"}
	case "NEIGHBORS_ONLY", "MUTUAL_ONLY":
		var relation string
		err := tx.QueryRow(ctx, `SELECT relation_type FROM follows
			WHERE follower_user_id = $1 AND target_blog_id = $2 AND status = 'ACCEPTED'`,
			userID, blogID).Scan(&relation)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &Fault{Status: http.StatusForbidden, Code: "COMMENTS_RESTRICTED", Message: "Only neighbors can comment on this post"}
		}
		if err != nil {
			return nil, err
		}
		if allowComments == "MUTUAL_ONLY" && relation != "MUTUAL_NEIGHBOR" {
			return nil, &Fault{Status: http.StatusForbidden, Code: "COMMENTS_RESTRICTED", Message: "Only mutual neighbors can comment on this post"}
		}
	case "REQUIRE_APPROVAL":
		commentStatus = "PENDING_APPROVAL"
	}

	depth := 0
	if input.ParentCommentID != nil && *input.ParentCommentID != "" {
		var parentDepth int
		var parentPost string
		err := tx.QueryRow(ctx, `SELECT depth, post_id FROM comments WHERE id = $1`, *input.ParentCommentID).
			Scan(&parentDepth, &parentPost)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &Fault{Status: http.StatusNotFound, Code: "PARENT_COMMENT_NOT_FOUND", Message: "Parent comment not found"}
		}
		if err != nil {
			return nil, err
		}
		if parentPost != postID {
			return nil, &Fault{Status: http.StatusBadRequest, Code: "PARENT_COMMENT_MISMATCH", Message: "Parent comment belongs to another post"}
		}
		if parentDepth != 0 {
			return nil, &Fault{Status: http.StatusBadRequest, Code: "COMMENT_DEPTH_EXCEEDED", Message: "Replies are limited to one level"}
		}
		depth = 1
	}

	var comment Comment
	err = tx.QueryRow(ctx, `
		INSERT INTO comments (post_id, blog_id, author_user_id, parent_comment_id, depth, content_text, is_secret, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, post_id, author_user_id, parent_comment_id, depth, content_text, is_secret, is_pinned, like_count, created_at
	`, postID, blogID, userID, input.ParentCommentID, depth, content, input.IsSecret, commentStatus).Scan(
		&comment.ID, &comment.PostID, &comment.AuthorUserID, &comment.ParentID,
		&comment.Depth, &comment.ContentText, &comment.IsSecret, &comment.IsPinned,
		&comment.LikeCount, &comment.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	if commentStatus == "PUBLISHED" {
		_, _ = tx.Exec(ctx, `UPDATE posts SET comment_count = comment_count + 1 WHERE id = $1`, postID)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	comment.AuthorName = s.displayName(ctx, userID)
	return &comment, nil
}

func (s *Service) displayName(ctx context.Context, userID string) string {
	var name string
	if err := s.pool.QueryRow(ctx, `SELECT display_name FROM user_profiles WHERE user_id = $1`, userID).Scan(&name); err == nil {
		return name
	}
	return ""
}

// ListComments returns published top-level comments with their replies.
// Secret comments are masked unless the viewer is the author, the commenter,
// or a blog member with the OWNER/ADMIN/EDITOR role.
func (s *Service) ListComments(ctx context.Context, viewerUserID, postID string, limit int) ([]Comment, error) {
	if err := s.requirePool(); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	var blogID string
	if err := s.pool.QueryRow(ctx, `SELECT blog_id FROM posts WHERE id = $1`, postID).Scan(&blogID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &Fault{Status: http.StatusNotFound, Code: "POST_NOT_FOUND", Message: "Post not found"}
		}
		return nil, err
	}
	canSeeSecrets := s.canModerate(ctx, viewerUserID, blogID)
	rows, err := s.pool.Query(ctx, `
		SELECT c.id, c.post_id, c.author_user_id, COALESCE(up.display_name, ''),
		       c.parent_comment_id, c.depth, c.content_text, c.is_secret, c.is_pinned,
		       c.like_count, c.created_at
		FROM comments c
		LEFT JOIN user_profiles up ON up.user_id = c.author_user_id
		WHERE c.post_id = $1 AND c.status = 'PUBLISHED'
		ORDER BY c.is_pinned DESC, c.created_at ASC
		LIMIT $2`, postID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	comments := make([]Comment, 0)
	for rows.Next() {
		var comment Comment
		if err := rows.Scan(&comment.ID, &comment.PostID, &comment.AuthorUserID, &comment.AuthorName,
			&comment.ParentID, &comment.Depth, &comment.ContentText, &comment.IsSecret,
			&comment.IsPinned, &comment.LikeCount, &comment.CreatedAt); err != nil {
			return nil, err
		}
		if comment.IsSecret && !canSeeSecrets && comment.AuthorUserID != viewerUserID {
			comment.ContentText = "비밀 댓글입니다."
		}
		comments = append(comments, comment)
	}
	return comments, rows.Err()
}

func (s *Service) canModerate(ctx context.Context, userID, blogID string) bool {
	if userID == "" {
		return false
	}
	var role string
	if err := s.pool.QueryRow(ctx, `SELECT role FROM blog_members WHERE blog_id = $1 AND user_id = $2`,
		blogID, userID).Scan(&role); err != nil {
		return false
	}
	return role == "OWNER" || role == "ADMIN" || role == "EDITOR"
}

// DeleteComment soft-deletes a comment owned by the user.
func (s *Service) DeleteComment(ctx context.Context, userID, commentID string) error {
	if err := s.requirePool(); err != nil {
		return err
	}
	tx, err := s.DB().Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var postID, status string
	var authorID string
	err = tx.QueryRow(ctx, `SELECT post_id, status, author_user_id FROM comments WHERE id = $1`, commentID).
		Scan(&postID, &status, &authorID)
	if errors.Is(err, pgx.ErrNoRows) {
		return &Fault{Status: http.StatusNotFound, Code: "COMMENT_NOT_FOUND", Message: "Comment not found"}
	}
	if err != nil {
		return err
	}
	if authorID != userID {
		return &Fault{Status: http.StatusForbidden, Code: "NOT_COMMENT_AUTHOR", Message: "Only the author can delete this comment"}
	}
	if _, err := tx.Exec(ctx, `UPDATE comments SET status = 'DELETED_BY_USER', updated_at = NOW() WHERE id = $1`, commentID); err != nil {
		return err
	}
	if status == "PUBLISHED" {
		_, _ = tx.Exec(ctx, `UPDATE posts SET comment_count = GREATEST(comment_count - 1, 0) WHERE id = $1`, postID)
	}
	return tx.Commit(ctx)
}

// Bookmark is one bookmarked post entry.
type Bookmark struct {
	PostID       string    `json:"post_id"`
	Title        string    `json:"title"`
	BlogHandle   string    `json:"blog_handle"`
	Slug         string    `json:"slug"`
	FolderName   string    `json:"folder_name"`
	ReadProgress int       `json:"read_progress_pct"`
	CreatedAt    time.Time `json:"created_at"`
}

// AddBookmark bookmarks a published post (idempotent).
func (s *Service) AddBookmark(ctx context.Context, userID, postID, folder string) error {
	if err := s.requirePool(); err != nil {
		return err
	}
	if folder == "" {
		folder = "기본 폴더"
	}
	tx, err := s.DB().Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var exists bool
	if err := tx.QueryRow(ctx, `SELECT TRUE FROM posts WHERE id = $1 AND status = 'PUBLISHED'`, postID).Scan(&exists); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &Fault{Status: http.StatusNotFound, Code: "POST_NOT_FOUND", Message: "Post not found"}
		}
		return err
	}
	var inserted bool
	err = tx.QueryRow(ctx, `INSERT INTO bookmarks (user_id, post_id, folder_name)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, post_id) DO UPDATE SET folder_name = EXCLUDED.folder_name
		RETURNING (xmax = 0)`, userID, postID, folder).Scan(&inserted)
	if err != nil {
		return err
	}
	if inserted {
		_, _ = tx.Exec(ctx, `UPDATE posts SET bookmark_count = bookmark_count + 1 WHERE id = $1`, postID)
	}
	return tx.Commit(ctx)
}

// RemoveBookmark removes a bookmark.
func (s *Service) RemoveBookmark(ctx context.Context, userID, postID string) error {
	if err := s.requirePool(); err != nil {
		return err
	}
	tx, err := s.DB().Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	res, err := tx.Exec(ctx, `DELETE FROM bookmarks WHERE user_id = $1 AND post_id = $2`, userID, postID)
	if err != nil {
		return err
	}
	if res.RowsAffected() > 0 {
		_, _ = tx.Exec(ctx, `UPDATE posts SET bookmark_count = GREATEST(bookmark_count - 1, 0) WHERE id = $1`, postID)
	}
	return tx.Commit(ctx)
}

// ListBookmarks returns the user's bookmarks newest-first.
func (s *Service) ListBookmarks(ctx context.Context, userID string, limit int) ([]Bookmark, error) {
	if err := s.requirePool(); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	rows, err := s.pool.Query(ctx, `
		SELECT bm.post_id, p.title, b.handle, p.slug, bm.folder_name, bm.read_progress_pct, bm.created_at
		FROM bookmarks bm
		JOIN posts p ON p.id = bm.post_id
		JOIN blogs b ON b.id = p.blog_id
		WHERE bm.user_id = $1
		ORDER BY bm.created_at DESC LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Bookmark, 0)
	for rows.Next() {
		var item Bookmark
		if err := rows.Scan(&item.PostID, &item.Title, &item.BlogHandle, &item.Slug, &item.FolderName, &item.ReadProgress, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// BlockUser blocks another user.
func (s *Service) BlockUser(ctx context.Context, userID, blockedID, reason string) error {
	if userID == blockedID {
		return &Fault{Status: http.StatusBadRequest, Code: "CANNOT_BLOCK_SELF", Message: "You cannot block yourself"}
	}
	if err := s.requirePool(); err != nil {
		return err
	}
	var exists bool
	if err := s.pool.QueryRow(ctx, `SELECT TRUE FROM users WHERE id = $1`, blockedID).Scan(&exists); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &Fault{Status: http.StatusNotFound, Code: "USER_NOT_FOUND", Message: "User not found"}
		}
		return err
	}
	_, err := s.pool.Exec(ctx, `INSERT INTO user_blocks (blocker_user_id, blocked_user_id, reason)
		VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`, userID, blockedID, reason)
	return err
}

// UnblockUser removes a block.
func (s *Service) UnblockUser(ctx context.Context, userID, blockedID string) error {
	if err := s.requirePool(); err != nil {
		return err
	}
	_, err := s.pool.Exec(ctx, `DELETE FROM user_blocks WHERE blocker_user_id = $1 AND blocked_user_id = $2`, userID, blockedID)
	return err
}
