package post

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Fault struct {
	Status  int
	Code    string
	Message string
}

func (f *Fault) Error() string {
	return f.Message
}

type PublishPostInput struct {
	Title             string          `json:"title"`
	Subtitle          string          `json:"subtitle,omitempty"`
	Slug              string          `json:"slug,omitempty"`
	CategoryID        *string         `json:"category_id,omitempty"`
	SeriesID          *string         `json:"series_id,omitempty"`
	SeriesOrder       *int            `json:"series_order,omitempty"`
	ContentJSON       json.RawMessage `json:"content_json"`
	ContentHTML       string          `json:"content_html"`
	ContentPlain      string          `json:"content_plain"`
	Excerpt           string          `json:"excerpt,omitempty"`
	CoverMediaID      *string         `json:"cover_media_id,omitempty"`
	Tags              []string        `json:"tags,omitempty"`
	Visibility        string          `json:"visibility"`
	AllowComments     string          `json:"allow_comments"`
	AllowReactions    *bool           `json:"allow_reactions,omitempty"`
	IsNotice          bool            `json:"is_notice"`
	IsPinned          bool            `json:"is_pinned"`
	SEOTitle          string          `json:"seo_title,omitempty"`
	SEODescription    string          `json:"seo_description,omitempty"`
	CanonicalURL      string          `json:"canonical_url,omitempty"`
	LicenseType       string          `json:"license_type,omitempty"`
	AIDisclosureLevel string          `json:"ai_disclosure_level,omitempty"`
	PaywallPriceMinor *int64          `json:"paywall_price_minor,omitempty"`
	ScheduledFor      *time.Time      `json:"scheduled_for,omitempty"`
}

type UpdatePostInput struct {
	Title             *string          `json:"title,omitempty"`
	Subtitle          *string          `json:"subtitle,omitempty"`
	Slug              *string          `json:"slug,omitempty"`
	CategoryID        *string          `json:"category_id,omitempty"`
	SeriesID          *string          `json:"series_id,omitempty"`
	SeriesOrder       *int             `json:"series_order,omitempty"`
	ContentJSON       *json.RawMessage `json:"content_json,omitempty"`
	ContentHTML       *string          `json:"content_html,omitempty"`
	ContentPlain      *string          `json:"content_plain,omitempty"`
	Excerpt           *string          `json:"excerpt,omitempty"`
	CoverMediaID      *string          `json:"cover_media_id,omitempty"`
	Tags              []string         `json:"tags,omitempty"`
	Visibility        *string          `json:"visibility,omitempty"`
	AllowComments     *string          `json:"allow_comments,omitempty"`
	AllowReactions    *bool            `json:"allow_reactions,omitempty"`
	IsNotice          *bool            `json:"is_notice,omitempty"`
	IsPinned          *bool            `json:"is_pinned,omitempty"`
	SEOTitle          *string          `json:"seo_title,omitempty"`
	SEODescription    *string          `json:"seo_description,omitempty"`
	CanonicalURL      *string          `json:"canonical_url,omitempty"`
	LicenseType       *string          `json:"license_type,omitempty"`
	AIDisclosureLevel *string          `json:"ai_disclosure_level,omitempty"`
	PaywallPriceMinor *int64           `json:"paywall_price_minor,omitempty"`
}

type PostDetail struct {
	ID                string          `json:"id"`
	BlogID            string          `json:"blog_id"`
	AuthorUserID      string          `json:"author_user_id"`
	AuthorName        string          `json:"author_name"`
	AuthorHandle      string          `json:"author_handle"`
	AuthorAvatarURL   *string         `json:"author_avatar_url,omitempty"`
	CategoryID        *string         `json:"category_id,omitempty"`
	CategoryName      *string         `json:"category_name,omitempty"`
	SeriesID          *string         `json:"series_id,omitempty"`
	SeriesTitle       *string         `json:"series_title,omitempty"`
	SeriesOrder       *int            `json:"series_order,omitempty"`
	Slug              string          `json:"slug"`
	Title             string          `json:"title"`
	Subtitle          string          `json:"subtitle"`
	ContentJSON       json.RawMessage `json:"content_json"`
	ContentHTML       string          `json:"content_html"`
	ContentPlain      string          `json:"content_plain"`
	Excerpt           string          `json:"excerpt"`
	CoverMediaID      *string         `json:"cover_media_id,omitempty"`
	CoverImageURL     *string         `json:"cover_image_url,omitempty"`
	Status            string          `json:"status"`
	Visibility        string          `json:"visibility"`
	AllowComments     string          `json:"allow_comments"`
	AllowReactions    bool            `json:"allow_reactions"`
	IsNotice          bool            `json:"is_notice"`
	IsPinned          bool            `json:"is_pinned"`
	SEOTitle          *string         `json:"seo_title,omitempty"`
	SEODescription    *string         `json:"seo_description,omitempty"`
	CanonicalURL      *string         `json:"canonical_url,omitempty"`
	LicenseType       string          `json:"license_type"`
	AIDisclosureLevel string          `json:"ai_disclosure_level"`
	PaywallPriceMinor *int64          `json:"paywall_price_minor,omitempty"`
	WordCount         int             `json:"word_count"`
	ReadingTimeSec    int             `json:"reading_time_sec"`
	ViewCount         int64           `json:"view_count"`
	ReactionCount     int             `json:"reaction_count"`
	CommentCount      int             `json:"comment_count"`
	BookmarkCount     int             `json:"bookmark_count"`
	RevisionNo        int             `json:"revision_no"`
	Tags              []string        `json:"tags"`
	IsPaywalled       bool            `json:"is_paywalled"`
	PublishedAt       *time.Time      `json:"published_at,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

type PostSummary struct {
	ID             string     `json:"id"`
	BlogID         string     `json:"blog_id"`
	AuthorUserID   string     `json:"author_user_id"`
	AuthorName     string     `json:"author_name"`
	Slug           string     `json:"slug"`
	Title          string     `json:"title"`
	Subtitle       string     `json:"subtitle"`
	Excerpt        string     `json:"excerpt"`
	CoverImageURL  *string    `json:"cover_image_url,omitempty"`
	CategoryName   *string    `json:"category_name,omitempty"`
	Status         string     `json:"status"`
	Visibility     string     `json:"visibility"`
	WordCount      int        `json:"word_count"`
	ReadingTimeSec int        `json:"reading_time_sec"`
	ViewCount      int64      `json:"view_count"`
	ReactionCount  int        `json:"reaction_count"`
	CommentCount   int        `json:"comment_count"`
	Tags           []string   `json:"tags"`
	PublishedAt    *time.Time `json:"published_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

var nonAlphaNumRegex = regexp.MustCompile(`[^\p{L}\p{N}]+`)

func Slugify(text string) string {
	slug := strings.ToLower(strings.TrimSpace(text))
	slug = nonAlphaNumRegex.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	if len(slug) > 180 {
		slug = slug[:180]
	}
	if slug == "" {
		slug = fmt.Sprintf("post-%d", time.Now().Unix())
	}
	return slug
}

func (s *Service) CheckAuthorRole(ctx context.Context, blogID, userID string) (string, error) {
	var role string
	err := s.pool.QueryRow(ctx, `SELECT role FROM blog_members WHERE blog_id = $1 AND user_id = $2`, blogID, userID).Scan(&role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", &Fault{Status: http.StatusForbidden, Code: "FORBIDDEN", Message: "Not a member of this blog"}
		}
		return "", err
	}
	switch role {
	case "OWNER", "ADMIN", "EDITOR", "AUTHOR":
		return role, nil
	default:
		return "", &Fault{Status: http.StatusForbidden, Code: "INSUFFICIENT_PERMISSIONS", Message: "Author privileges or above required"}
	}
}

func (s *Service) PublishPost(ctx context.Context, blogID, postID, userID string, input PublishPostInput) (*PostDetail, error) {
	if s.pool != nil {
		role, err := s.CheckAuthorRole(ctx, blogID, userID)
		if err != nil {
			return nil, err
		}
		_ = role
	}

	title := strings.TrimSpace(input.Title)
	if title == "" {
		return nil, &Fault{Status: http.StatusBadRequest, Code: "TITLE_REQUIRED", Message: "Title is required"}
	}

	slug := input.Slug
	if slug == "" {
		slug = Slugify(title)
	}

	visibility := strings.ToUpper(strings.TrimSpace(input.Visibility))
	if visibility == "" {
		visibility = "PUBLIC"
	}
	switch visibility {
	case "PUBLIC", "UNLISTED", "NEIGHBOR_ONLY", "MUTUAL_ONLY", "PAID_MEMBERS_ONLY", "PRIVATE":
	default:
		visibility = "PUBLIC"
	}

	allowComments := strings.ToUpper(strings.TrimSpace(input.AllowComments))
	if allowComments == "" {
		allowComments = "ALL_ALLOWED"
	}

	allowReactions := true
	if input.AllowReactions != nil {
		allowReactions = *input.AllowReactions
	}

	words := len(strings.Fields(input.ContentPlain))
	readingTime := (words / 200) * 60
	if readingTime < 30 && words > 0 {
		readingTime = 30
	}

	excerpt := input.Excerpt
	if excerpt == "" && len(input.ContentPlain) > 0 {
		runes := []rune(input.ContentPlain)
		if len(runes) > 160 {
			excerpt = string(runes[:160]) + "..."
		} else {
			excerpt = string(runes)
		}
	}

	seoDesc := input.SEODescription
	if seoDesc == "" {
		seoDesc = excerpt
	}

	status := "PUBLISHED"
	now := time.Now().UTC()
	var publishedAt *time.Time = &now

	if input.ScheduledFor != nil && input.ScheduledFor.After(now) {
		status = "SCHEDULED"
		publishedAt = nil
	}

	if s.pool == nil {
		return &PostDetail{
			ID:             postID,
			BlogID:         blogID,
			Title:          title,
			Slug:           slug,
			Status:         status,
			Visibility:     visibility,
			WordCount:      words,
			ReadingTimeSec: readingTime,
			PublishedAt:    publishedAt,
		}, nil
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Check existing post and unique slug
	var existingSlug string
	var existingStatus string
	var currentRevNo int
	err = tx.QueryRow(ctx, `SELECT slug, status, revision_no FROM posts WHERE id = $1 AND blog_id = $2`, postID, blogID).Scan(&existingSlug, &existingStatus, &currentRevNo)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &Fault{Status: http.StatusNotFound, Code: "POST_NOT_FOUND", Message: "Post not found"}
		}
		return nil, err
	}

	// Slug conflict check with other posts
	var conflictID string
	err = tx.QueryRow(ctx, `SELECT id FROM posts WHERE blog_id = $1 AND slug = $2 AND id <> $3`, blogID, slug, postID).Scan(&conflictID)
	if err == nil {
		// Auto append suffix
		slug = fmt.Sprintf("%s-%d", slug, time.Now().Unix()%10000)
	}

	// Update slug history if changed and previously published
	if existingSlug != slug && existingStatus == "PUBLISHED" {
		_, _ = tx.Exec(ctx, `INSERT INTO post_slug_history (blog_id, post_id, old_slug) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`, blogID, postID, existingSlug)
	}

	newRevNo := currentRevNo + 1

	// Update posts record
	_, err = tx.Exec(ctx, `
		UPDATE posts
		SET title = $1, subtitle = $2, slug = $3, category_id = $4, series_id = $5, series_order = $6,
		    content_json = $7, content_html = $8, content_plain = $9, excerpt = $10,
		    cover_media_id = $11, status = $12, visibility = $13, allow_comments = $14,
		    allow_reactions = $15, is_notice = $16, is_pinned = $17, seo_title = $18,
		    seo_description = $19, canonical_url = $20, license_type = COALESCE(NULLIF($21, ''), 'ALL_RIGHTS_RESERVED'),
		    ai_disclosure_level = COALESCE(NULLIF($22, ''), 'NONE'), paywall_price_minor = $23,
		    word_count = $24, reading_time_sec = $25, revision_no = $26,
		    published_at = COALESCE(published_at, $27), updated_at = NOW()
		WHERE id = $28 AND blog_id = $29
	`, title, input.Subtitle, slug, input.CategoryID, input.SeriesID, input.SeriesOrder,
		input.ContentJSON, input.ContentHTML, input.ContentPlain, excerpt,
		input.CoverMediaID, status, visibility, allowComments,
		allowReactions, input.IsNotice, input.IsPinned, input.SEOTitle,
		seoDesc, input.CanonicalURL, input.LicenseType,
		input.AIDisclosureLevel, input.PaywallPriceMinor,
		words, readingTime, newRevNo, publishedAt, postID, blogID,
	)
	if err != nil {
		return nil, err
	}

	// If scheduled, insert/update scheduled_publications
	if status == "SCHEDULED" {
		_, err = tx.Exec(ctx, `
			INSERT INTO scheduled_publications (post_id, blog_id, scheduled_by, scheduled_for, status)
			VALUES ($1, $2, $3, $4, 'PENDING')
			ON CONFLICT (post_id) DO UPDATE SET
				scheduled_for = EXCLUDED.scheduled_for,
				status = 'PENDING',
				updated_at = NOW()
		`, postID, blogID, userID, input.ScheduledFor)
		if err != nil {
			return nil, err
		}
	} else if status == "PUBLISHED" {
		// Create published revision snapshot
		_, err = tx.Exec(ctx, `
			INSERT INTO post_revisions (
				post_id, revision_no, editor_user_id, title, subtitle,
				content_json, content_html, change_summary, is_published_snapshot
			) VALUES ($1, $2, $3, $4, $5, $6, $7, 'Published snapshot', TRUE)
			ON CONFLICT (post_id, revision_no) DO NOTHING
		`, postID, newRevNo, userID, title, input.Subtitle, input.ContentJSON, input.ContentHTML)
		if err != nil {
			return nil, err
		}

		// Enqueue search index job
		_, err = tx.Exec(ctx, `
			INSERT INTO search_index_jobs (target_index, document_id, operation, priority)
			VALUES ('POSTS', $1, 'UPSERT', 80)
		`, postID)
		if err != nil {
			return nil, err
		}

		// Outbox event
		payload, _ := json.Marshal(map[string]any{
			"post_id":      postID,
			"blog_id":      blogID,
			"slug":         slug,
			"title":        title,
			"published_at": now.Format(time.RFC3339),
		})
		_, err = tx.Exec(ctx, `
			INSERT INTO outbox_events (aggregate_type, aggregate_id, event_type, payload_json, idempotency_key)
			VALUES ('POST', $1, 'post.published', $2, 'evt-post-pub-' || $1 || '-' || $3)
		`, postID, payload, newRevNo)
		if err != nil {
			return nil, err
		}
	}

	// Tags handling
	if len(input.Tags) > 0 {
		_, _ = tx.Exec(ctx, `DELETE FROM post_tags WHERE post_id = $1`, postID)
		for i, rawTag := range input.Tags {
			tagName := strings.TrimSpace(rawTag)
			if tagName == "" {
				continue
			}
			tagSlug := Slugify(tagName)
			var tagID string
			err = tx.QueryRow(ctx, `
				INSERT INTO tags (name, slug, post_count)
				VALUES ($1, $2, 1)
				ON CONFLICT (slug) DO UPDATE SET post_count = tags.post_count + 1
				RETURNING id
			`, tagName, tagSlug).Scan(&tagID)
			if err == nil {
				_, _ = tx.Exec(ctx, `
					INSERT INTO post_tags (post_id, tag_id, sort_order)
					VALUES ($1, $2, $3)
					ON CONFLICT DO NOTHING
				`, postID, tagID, i)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &PostDetail{
		ID:                postID,
		BlogID:            blogID,
		AuthorUserID:      userID,
		Title:             title,
		Subtitle:          input.Subtitle,
		Slug:              slug,
		Status:            status,
		Visibility:        visibility,
		AllowComments:     allowComments,
		AllowReactions:    allowReactions,
		IsNotice:          input.IsNotice,
		IsPinned:          input.IsPinned,
		WordCount:         words,
		ReadingTimeSec:    readingTime,
		RevisionNo:        newRevNo,
		PublishedAt:       publishedAt,
		Tags:              input.Tags,
		PaywallPriceMinor: input.PaywallPriceMinor,
		ContentJSON:       input.ContentJSON,
		ContentHTML:       input.ContentHTML,
		ContentPlain:      input.ContentPlain,
	}, nil
}

func (s *Service) UnpublishPost(ctx context.Context, blogID, postID, userID string) error {
	if s.pool != nil {
		if _, err := s.CheckAuthorRole(ctx, blogID, userID); err != nil {
			return err
		}
	}
	if s.pool == nil {
		return nil
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	res, err := tx.Exec(ctx, `
		UPDATE posts SET status = 'UNPUBLISHED', updated_at = NOW()
		WHERE id = $1 AND blog_id = $2
	`, postID, blogID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return &Fault{Status: http.StatusNotFound, Code: "POST_NOT_FOUND", Message: "Post not found"}
	}

	// Delete from search index
	_, _ = tx.Exec(ctx, `INSERT INTO search_index_jobs (target_index, document_id, operation, priority) VALUES ('POSTS', $1, 'DELETE', 90)`, postID)

	// Outbox event
	payload, _ := json.Marshal(map[string]any{"post_id": postID, "blog_id": blogID})
	_, _ = tx.Exec(ctx, `
		INSERT INTO outbox_events (aggregate_type, aggregate_id, event_type, payload_json, idempotency_key)
		VALUES ('POST', $1, 'post.unpublished', $2, 'evt-post-unpub-' || $1 || '-' || extract(epoch from now()))
	`, postID, payload)

	return tx.Commit(ctx)
}

func (s *Service) TrashPost(ctx context.Context, blogID, postID, userID string) error {
	if s.pool != nil {
		if _, err := s.CheckAuthorRole(ctx, blogID, userID); err != nil {
			return err
		}
	}
	if s.pool == nil {
		return nil
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	now := time.Now().UTC()
	purgeAfter := now.Add(30 * 24 * time.Hour)

	res, err := tx.Exec(ctx, `
		UPDATE posts
		SET status = 'TRASHED', trashed_at = $1, purge_after_at = $2, updated_at = NOW()
		WHERE id = $3 AND blog_id = $4
	`, now, purgeAfter, postID, blogID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return &Fault{Status: http.StatusNotFound, Code: "POST_NOT_FOUND", Message: "Post not found"}
	}

	_, _ = tx.Exec(ctx, `INSERT INTO search_index_jobs (target_index, document_id, operation, priority) VALUES ('POSTS', $1, 'DELETE', 90)`, postID)

	return tx.Commit(ctx)
}

func (s *Service) RestorePost(ctx context.Context, blogID, postID, userID string) error {
	if s.pool != nil {
		if _, err := s.CheckAuthorRole(ctx, blogID, userID); err != nil {
			return err
		}
	}
	if s.pool == nil {
		return nil
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	res, err := tx.Exec(ctx, `
		UPDATE posts
		SET status = 'PUBLISHED', trashed_at = NULL, purge_after_at = NULL, updated_at = NOW()
		WHERE id = $1 AND blog_id = $2 AND status = 'TRASHED'
	`, postID, blogID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return &Fault{Status: http.StatusNotFound, Code: "POST_NOT_IN_TRASH", Message: "Post is not in trash"}
	}

	_, _ = tx.Exec(ctx, `INSERT INTO search_index_jobs (target_index, document_id, operation, priority) VALUES ('POSTS', $1, 'UPSERT', 80)`, postID)

	return tx.Commit(ctx)
}

func (s *Service) GetPublicPost(ctx context.Context, blogKey, slug string, viewerUserID string) (*PostDetail, error) {
	if s.pool == nil {
		return &PostDetail{Title: "Mock Post", Slug: slug, Status: "PUBLISHED"}, nil
	}

	// Resolve blogKey (could be UUID or handle)
	var blogID string
	err := s.pool.QueryRow(ctx, `SELECT id FROM blogs WHERE id::text = $1 OR handle = $1`, blogKey).Scan(&blogID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &Fault{Status: http.StatusNotFound, Code: "BLOG_NOT_FOUND", Message: "Blog not found"}
		}
		return nil, err
	}

	var p PostDetail
	var categoryName, seriesTitle, authorAvatar *string
	err = s.pool.QueryRow(ctx, `
		SELECT p.id, p.blog_id, p.author_user_id, u.nickname, u.handle, up.avatar_url,
		       p.category_id, bc.name, p.series_id, s.title, p.series_order,
		       p.slug, p.title, p.subtitle, p.content_json, p.content_html, p.content_plain,
		       p.excerpt, p.cover_media_id, p.status, p.visibility, p.allow_comments,
		       p.allow_reactions, p.is_notice, p.is_pinned, p.seo_title, p.seo_description,
		       p.canonical_url, p.license_type, p.ai_disclosure_level, p.paywall_price_minor,
		       p.word_count, p.reading_time_sec, p.view_count, p.reaction_count, p.comment_count,
		       p.bookmark_count, p.revision_no, p.published_at, p.created_at, p.updated_at
		FROM posts p
		JOIN users u ON u.id = p.author_user_id
		LEFT JOIN user_profiles up ON up.user_id = u.id
		LEFT JOIN blog_categories bc ON bc.id = p.category_id
		LEFT JOIN series s ON s.id = p.series_id
		WHERE p.blog_id = $1 AND p.slug = $2 AND p.status = 'PUBLISHED' AND p.trashed_at IS NULL
	`, blogID, slug).Scan(
		&p.ID, &p.BlogID, &p.AuthorUserID, &p.AuthorName, &p.AuthorHandle, &authorAvatar,
		&p.CategoryID, &categoryName, &p.SeriesID, &seriesTitle, &p.SeriesOrder,
		&p.Slug, &p.Title, &p.Subtitle, &p.ContentJSON, &p.ContentHTML, &p.ContentPlain,
		&p.Excerpt, &p.CoverMediaID, &p.Status, &p.Visibility, &p.AllowComments,
		&p.AllowReactions, &p.IsNotice, &p.IsPinned, &p.SEOTitle, &p.SEODescription,
		&p.CanonicalURL, &p.LicenseType, &p.AIDisclosureLevel, &p.PaywallPriceMinor,
		&p.WordCount, &p.ReadingTimeSec, &p.ViewCount, &p.ReactionCount, &p.CommentCount,
		&p.BookmarkCount, &p.RevisionNo, &p.PublishedAt, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Check if slug exists in post_slug_history
			var oldPostID string
			if hErr := s.pool.QueryRow(ctx, `SELECT post_id FROM post_slug_history WHERE blog_id = $1 AND old_slug = $2`, blogID, slug).Scan(&oldPostID); hErr == nil {
				var currentSlug string
				_ = s.pool.QueryRow(ctx, `SELECT slug FROM posts WHERE id = $1`, oldPostID).Scan(&currentSlug)
				return nil, &Fault{Status: http.StatusMovedPermanently, Code: "MOVED_PERMANENTLY", Message: currentSlug}
			}
			return nil, &Fault{Status: http.StatusNotFound, Code: "POST_NOT_FOUND", Message: "Post not found or unpublished"}
		}
		return nil, err
	}

	p.CategoryName = categoryName
	p.SeriesTitle = seriesTitle
	p.AuthorAvatarURL = authorAvatar

	// Visibility permission check
	if p.Visibility == "PRIVATE" && viewerUserID != p.AuthorUserID {
		return nil, &Fault{Status: http.StatusForbidden, Code: "POST_PRIVATE", Message: "This post is private"}
	}

	// Paywall check: if paywall price > 0, check if purchased
	if p.PaywallPriceMinor != nil && *p.PaywallPriceMinor > 0 && viewerUserID != p.AuthorUserID {
		p.IsPaywalled = true
		// Mask HTML content
		runes := []rune(p.ContentPlain)
		previewLength := 200
		if len(runes) < previewLength {
			previewLength = len(runes)
		}
		p.ContentHTML = fmt.Sprintf("<div class='paywall-preview'><p>%s</p><div class='paywall-notice'>유료 멤버십 또는 단건 결제 후 열람 가능한 콘텐츠입니다.</div></div>", string(runes[:previewLength]))
		p.ContentJSON = json.RawMessage(`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"유료 콘텐츠입니다."}]}]}`)
	}

	// Increment view count asynchronously
	go func(postID string) {
		_, _ = s.pool.Exec(context.Background(), `UPDATE posts SET view_count = view_count + 1 WHERE id = $1`, postID)
	}(p.ID)

	// Fetch tags
	tagRows, _ := s.pool.Query(ctx, `
		SELECT t.name FROM tags t
		JOIN post_tags pt ON pt.tag_id = t.id
		WHERE pt.post_id = $1
		ORDER BY pt.sort_order ASC
	`, p.ID)
	if tagRows != nil {
		defer tagRows.Close()
		p.Tags = make([]string, 0)
		for tagRows.Next() {
			var tName string
			if err := tagRows.Scan(&tName); err == nil {
				p.Tags = append(p.Tags, tName)
			}
		}
	}

	return &p, nil
}

func (s *Service) ListPublicPosts(ctx context.Context, blogKey string, categoryID, tag string, limit int, cursor string) ([]PostSummary, error) {
	if s.pool == nil {
		return []PostSummary{}, nil
	}

	var blogID string
	err := s.pool.QueryRow(ctx, `SELECT id FROM blogs WHERE id::text = $1 OR handle = $1`, blogKey).Scan(&blogID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &Fault{Status: http.StatusNotFound, Code: "BLOG_NOT_FOUND", Message: "Blog not found"}
		}
		return nil, err
	}

	if limit <= 0 || limit > 50 {
		limit = 10
	}

	query := `
		SELECT p.id, p.blog_id, p.author_user_id, u.nickname, p.slug, p.title, p.subtitle,
		       p.excerpt, bc.name, p.status, p.visibility, p.word_count, p.reading_time_sec,
		       p.view_count, p.reaction_count, p.comment_count, p.published_at, p.created_at
		FROM posts p
		JOIN users u ON u.id = p.author_user_id
		LEFT JOIN blog_categories bc ON bc.id = p.category_id
		WHERE p.blog_id = $1 AND p.status = 'PUBLISHED' AND p.visibility = 'PUBLIC' AND p.trashed_at IS NULL
	`
	args := []any{blogID}
	argIdx := 2

	if categoryID != "" {
		query += fmt.Sprintf(" AND p.category_id = $%d", argIdx)
		args = append(args, categoryID)
		argIdx++
	}

	query += fmt.Sprintf(" ORDER BY p.is_pinned DESC, p.published_at DESC LIMIT $%d", argIdx)
	args = append(args, limit)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]PostSummary, 0)
	for rows.Next() {
		var s PostSummary
		var catName *string
		if err := rows.Scan(
			&s.ID, &s.BlogID, &s.AuthorUserID, &s.AuthorName, &s.Slug, &s.Title, &s.Subtitle,
			&s.Excerpt, &catName, &s.Status, &s.Visibility, &s.WordCount, &s.ReadingTimeSec,
			&s.ViewCount, &s.ReactionCount, &s.CommentCount, &s.PublishedAt, &s.CreatedAt,
		); err != nil {
			return nil, err
		}
		s.CategoryName = catName
		list = append(list, s)
	}
	return list, rows.Err()
}
