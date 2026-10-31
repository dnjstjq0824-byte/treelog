package editor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"treeblog/internal/pkg/safenet"
)

type Fault struct {
	Status  int
	Code    string
	Message string
	Data    any
}

func (f *Fault) Error() string {
	return f.Message
}

type CreateDraftInput struct {
	Title        string          `json:"title"`
	Subtitle     string          `json:"subtitle,omitempty"`
	CategoryID   *string         `json:"category_id,omitempty"`
	ContentJSON  json.RawMessage `json:"content_json,omitempty"`
	ContentHTML  string          `json:"content_html,omitempty"`
	ContentPlain string          `json:"content_plain,omitempty"`
}

type AutosaveInput struct {
	BaseRevisionNo int             `json:"base_revision_no"`
	Title          string          `json:"title"`
	Subtitle       string          `json:"subtitle,omitempty"`
	ContentJSON    json.RawMessage `json:"content_json"`
	ClientTime     time.Time       `json:"client_time"`
}

type LinkPreviewInput struct {
	URL string `json:"url"`
}

type LinkPreviewResponse struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
	SiteName    string `json:"site_name"`
}

type MarkdownConvertInput struct {
	Direction string `json:"direction"` // "TO_AST" or "TO_MARKDOWN"
	Markdown  string `json:"markdown,omitempty"`
	AST       string `json:"ast,omitempty"`
}

type RevisionSummary struct {
	ID                  string    `json:"id"`
	PostID              string    `json:"post_id"`
	RevisionNo          int       `json:"revision_no"`
	EditorUserID        string    `json:"editor_user_id"`
	Title               string    `json:"title"`
	Subtitle            string    `json:"subtitle"`
	ChangeSummary       string    `json:"change_summary"`
	IsPublishedSnapshot bool      `json:"is_published_snapshot"`
	CreatedAt           time.Time `json:"created_at"`
}

type RevisionDetail struct {
	RevisionSummary
	ContentJSON json.RawMessage `json:"content_json"`
	ContentHTML string          `json:"content_html"`
}

type Service struct {
	pool       *pgxpool.Pool
	safeClient *http.Client
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{
		pool:       pool,
		safeClient: safenet.NewSafeHTTPClient(8 * time.Second),
	}
}

func (s *Service) CheckAuthorRole(ctx context.Context, blogID, userID string) error {
	var role string
	err := s.pool.QueryRow(ctx, `SELECT role FROM blog_members WHERE blog_id = $1 AND user_id = $2`, blogID, userID).Scan(&role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &Fault{Status: http.StatusForbidden, Code: "FORBIDDEN", Message: "Not a member of this blog"}
		}
		return err
	}
	switch role {
	case "OWNER", "ADMIN", "EDITOR", "AUTHOR":
		return nil
	default:
		return &Fault{Status: http.StatusForbidden, Code: "INSUFFICIENT_PERMISSIONS", Message: "Author privileges or above required"}
	}
}

func (s *Service) CreateDraft(ctx context.Context, blogID, userID string, input CreateDraftInput) (map[string]any, error) {
	if s.pool != nil {
		if err := s.CheckAuthorRole(ctx, blogID, userID); err != nil {
			return nil, err
		}
	}

	title := strings.TrimSpace(input.Title)
	if title == "" {
		title = "제목 없는 글"
	}

	contentJSON := input.ContentJSON
	if len(contentJSON) == 0 {
		contentJSON = json.RawMessage(`{"type":"doc","content":[]}`)
	}

	// Generate initial temporary slug
	slug := fmt.Sprintf("draft-%d", time.Now().UnixNano())

	if s.pool == nil {
		return map[string]any{
			"id":          "draft-mock-id",
			"blog_id":     blogID,
			"title":       title,
			"status":      "DRAFT",
			"revision_no": 1,
		}, nil
	}

	var postID string
	var revNo int
	err := s.pool.QueryRow(ctx, `
		INSERT INTO posts (
			blog_id, author_user_id, category_id, slug, title, subtitle,
			content_json, content_html, content_plain, status, revision_no
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 'DRAFT', 1)
		RETURNING id, revision_no
	`, blogID, userID, input.CategoryID, slug, title, input.Subtitle,
		contentJSON, input.ContentHTML, input.ContentPlain,
	).Scan(&postID, &revNo)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"id":          postID,
		"blog_id":     blogID,
		"title":       title,
		"subtitle":    input.Subtitle,
		"status":      "DRAFT",
		"revision_no": revNo,
		"slug":        slug,
	}, nil
}

func (s *Service) AutosaveDraft(ctx context.Context, blogID, postID, userID string, input AutosaveInput) (map[string]any, error) {
	if s.pool != nil {
		if err := s.CheckAuthorRole(ctx, blogID, userID); err != nil {
			return nil, err
		}
	}

	if s.pool == nil {
		return map[string]any{"saved": true, "post_id": postID}, nil
	}

	// Check post exists and check optimistic locking
	var currentRevNo int
	var currentTitle, currentSubtitle string
	var currentJSON json.RawMessage
	err := s.pool.QueryRow(ctx, `
		SELECT revision_no, title, subtitle, content_json
		FROM posts
		WHERE id = $1 AND blog_id = $2
	`, postID, blogID).Scan(&currentRevNo, &currentTitle, &currentSubtitle, &currentJSON)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &Fault{Status: http.StatusNotFound, Code: "POST_NOT_FOUND", Message: "Post not found"}
		}
		return nil, err
	}

	// If client's base revision is behind current server revision -> conflict!
	if input.BaseRevisionNo > 0 && input.BaseRevisionNo < currentRevNo {
		return nil, &Fault{
			Status:  http.StatusConflict,
			Code:    "EDIT_VERSION_CONFLICT",
			Message: "A newer version exists on the server. Please resolve conflicts.",
			Data: map[string]any{
				"server_revision_no": currentRevNo,
				"server_title":       currentTitle,
				"server_subtitle":    currentSubtitle,
				"server_content":     currentJSON,
			},
		}
	}

	now := time.Now().UTC()
	clientTime := input.ClientTime
	if clientTime.IsZero() {
		clientTime = now
	}

	// Upsert into autosave_drafts
	_, err = s.pool.Exec(ctx, `
		INSERT INTO autosave_drafts (
			post_id, user_id, base_revision_no, title, subtitle, content_json, client_updated_at, server_saved_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (post_id) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			base_revision_no = EXCLUDED.base_revision_no,
			title = EXCLUDED.title,
			subtitle = EXCLUDED.subtitle,
			content_json = EXCLUDED.content_json,
			client_updated_at = EXCLUDED.client_updated_at,
			server_saved_at = EXCLUDED.server_saved_at
	`, postID, userID, input.BaseRevisionNo, input.Title, input.Subtitle, input.ContentJSON, clientTime, now)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"saved":           true,
		"post_id":         postID,
		"server_saved_at": now.Format(time.RFC3339),
	}, nil
}

func (s *Service) ListRevisions(ctx context.Context, blogID, postID, userID string) ([]RevisionSummary, error) {
	if s.pool != nil {
		if err := s.CheckAuthorRole(ctx, blogID, userID); err != nil {
			return nil, err
		}
	}

	if s.pool == nil {
		return []RevisionSummary{}, nil
	}

	rows, err := s.pool.Query(ctx, `
		SELECT id, post_id, revision_no, editor_user_id, title, subtitle,
		       change_summary, is_published_snapshot, created_at
		FROM post_revisions
		WHERE post_id = $1
		ORDER BY revision_no DESC
	`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]RevisionSummary, 0)
	for rows.Next() {
		var r RevisionSummary
		if err := rows.Scan(
			&r.ID, &r.PostID, &r.RevisionNo, &r.EditorUserID, &r.Title, &r.Subtitle,
			&r.ChangeSummary, &r.IsPublishedSnapshot, &r.CreatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, rows.Err()
}

func (s *Service) RestoreRevision(ctx context.Context, blogID, postID, revisionID, userID string) (*RevisionDetail, error) {
	if s.pool != nil {
		if err := s.CheckAuthorRole(ctx, blogID, userID); err != nil {
			return nil, err
		}
	}

	if s.pool == nil {
		return &RevisionDetail{}, nil
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Fetch target revision
	var target RevisionDetail
	err = tx.QueryRow(ctx, `
		SELECT id, post_id, revision_no, editor_user_id, title, subtitle,
		       content_json, content_html, change_summary, is_published_snapshot, created_at
		FROM post_revisions
		WHERE id = $1 AND post_id = $2
	`, revisionID, postID).Scan(
		&target.ID, &target.PostID, &target.RevisionNo, &target.EditorUserID, &target.Title, &target.Subtitle,
		&target.ContentJSON, &target.ContentHTML, &target.ChangeSummary, &target.IsPublishedSnapshot, &target.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &Fault{Status: http.StatusNotFound, Code: "REVISION_NOT_FOUND", Message: "Revision not found"}
		}
		return nil, err
	}

	// Increment post revision_no
	var newRevNo int
	err = tx.QueryRow(ctx, `
		UPDATE posts
		SET revision_no = revision_no + 1,
		    title = $1,
		    subtitle = $2,
		    content_json = $3,
		    content_html = $4,
		    updated_at = NOW()
		WHERE id = $5 AND blog_id = $6
		RETURNING revision_no
	`, target.Title, target.Subtitle, target.ContentJSON, target.ContentHTML, postID, blogID).Scan(&newRevNo)
	if err != nil {
		return nil, err
	}

	// Create non-destructive new revision snapshot
	summary := fmt.Sprintf("Restored from revision #%d", target.RevisionNo)
	var created RevisionDetail
	err = tx.QueryRow(ctx, `
		INSERT INTO post_revisions (
			post_id, revision_no, editor_user_id, title, subtitle,
			content_json, content_html, change_summary, is_published_snapshot
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, FALSE)
		RETURNING id, post_id, revision_no, editor_user_id, title, subtitle,
		          content_json, content_html, change_summary, is_published_snapshot, created_at
	`, postID, newRevNo, userID, target.Title, target.Subtitle,
		target.ContentJSON, target.ContentHTML, summary,
	).Scan(
		&created.ID, &created.PostID, &created.RevisionNo, &created.EditorUserID, &created.Title, &created.Subtitle,
		&created.ContentJSON, &created.ContentHTML, &created.ChangeSummary, &created.IsPublishedSnapshot, &created.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &created, nil
}

var (
	titleTagRegex   = regexp.MustCompile(`(?i)<title[^>]*>([^<]+)</title>`)
	ogTitleRegex    = regexp.MustCompile(`(?i)<meta\s+property=["']og:title["']\s+content=["']([^"']+)["']`)
	ogDescRegex     = regexp.MustCompile(`(?i)<meta\s+property=["']og:description["']\s+content=["']([^"']+)["']`)
	ogImageRegex    = regexp.MustCompile(`(?i)<meta\s+property=["']og:image["']\s+content=["']([^"']+)["']`)
	ogSiteNameRegex = regexp.MustCompile(`(?i)<meta\s+property=["']og:site_name["']\s+content=["']([^"']+)["']`)
)

func (s *Service) FetchLinkPreview(ctx context.Context, rawURL string) (*LinkPreviewResponse, error) {
	rawURL = strings.TrimSpace(rawURL)
	parsed, err := url.Parse(rawURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return nil, &Fault{Status: http.StatusBadRequest, Code: "INVALID_URL", Message: "Valid HTTP/HTTPS URL is required"}
	}

	client := s.safeClient
	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; TreeBlogBot/1.0; +https://treeblog.net/bot)")

	resp, err := client.Do(req)
	if err != nil {
		return nil, &Fault{Status: http.StatusBadGateway, Code: "LINK_FETCH_FAILED", Message: fmt.Sprintf("Failed to fetch target URL: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, &Fault{Status: http.StatusBadGateway, Code: "LINK_FETCH_ERROR", Message: fmt.Sprintf("Target server responded with HTTP %d", resp.StatusCode)}
	}

	// Limit read to 256KB to avoid memory exhaustion
	limitReader := io.LimitReader(resp.Body, 256*1024)
	bodyBytes, err := io.ReadAll(limitReader)
	if err != nil {
		return nil, err
	}
	htmlStr := string(bodyBytes)

	preview := &LinkPreviewResponse{
		URL:      rawURL,
		SiteName: parsed.Host,
	}

	if match := ogTitleRegex.FindStringSubmatch(htmlStr); len(match) > 1 {
		preview.Title = strings.TrimSpace(match[1])
	} else if match := titleTagRegex.FindStringSubmatch(htmlStr); len(match) > 1 {
		preview.Title = strings.TrimSpace(match[1])
	}

	if match := ogDescRegex.FindStringSubmatch(htmlStr); len(match) > 1 {
		preview.Description = strings.TrimSpace(match[1])
	}

	if match := ogImageRegex.FindStringSubmatch(htmlStr); len(match) > 1 {
		imgURL := strings.TrimSpace(match[1])
		if strings.HasPrefix(imgURL, "//") {
			imgURL = parsed.Scheme + ":" + imgURL
		} else if strings.HasPrefix(imgURL, "/") {
			imgURL = parsed.Scheme + "://" + parsed.Host + imgURL
		}
		preview.ImageURL = imgURL
	}

	if match := ogSiteNameRegex.FindStringSubmatch(htmlStr); len(match) > 1 {
		preview.SiteName = strings.TrimSpace(match[1])
	}

	return preview, nil
}

func (s *Service) ConvertMarkdown(ctx context.Context, input MarkdownConvertInput) (map[string]any, error) {
	if strings.ToUpper(input.Direction) == "TO_MARKDOWN" {
		// Mock/Basic AST to Markdown conversion
		return map[string]any{
			"markdown": "# Converted Document\n\nContent rendered from Tiptap AST.",
		}, nil
	}

	// Default: TO_AST (Markdown to Tiptap JSON AST)
	lines := strings.Split(input.Markdown, "\n")
	contentNodes := make([]map[string]any, 0)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "# ") {
			contentNodes = append(contentNodes, map[string]any{
				"type":  "heading",
				"attrs": map[string]any{"level": 1},
				"content": []map[string]any{
					{"type": "text", "text": strings.TrimPrefix(trimmed, "# ")},
				},
			})
		} else if strings.HasPrefix(trimmed, "## ") {
			contentNodes = append(contentNodes, map[string]any{
				"type":  "heading",
				"attrs": map[string]any{"level": 2},
				"content": []map[string]any{
					{"type": "text", "text": strings.TrimPrefix(trimmed, "## ")},
				},
			})
		} else {
			contentNodes = append(contentNodes, map[string]any{
				"type": "paragraph",
				"content": []map[string]any{
					{"type": "text", "text": trimmed},
				},
			})
		}
	}

	astDoc := map[string]any{
		"type":    "doc",
		"content": contentNodes,
	}

	return map[string]any{
		"ast": astDoc,
	}, nil
}
