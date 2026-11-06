package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"treeblog/internal/modules/media"
	"treeblog/internal/modules/treescan"
	"treeblog/internal/platform/opensearch"
)

// Dependencies bundles the external systems event handlers need.
type Dependencies struct {
	Pool       *pgxpool.Pool
	Logger     *slog.Logger
	OpenSearch *opensearch.Client
	Media      *media.Processor
	Treescan   *treescan.Service
	AppBaseURL string
}

// DefaultHandlers returns handlers for every outbox event type the API emits.
func DefaultHandlers(deps Dependencies) map[string]Handler {
	return map[string]Handler{
		"system.ping": func(context.Context, json.RawMessage) error { return nil },
		"post.published": func(ctx context.Context, payload json.RawMessage) error {
			return handlePostPublished(ctx, deps, payload)
		},
		"post.unpublished": func(ctx context.Context, payload json.RawMessage) error {
			return handlePostUnpublished(ctx, deps, payload)
		},
		"media.process": func(ctx context.Context, payload json.RawMessage) error {
			return handleMediaProcess(ctx, deps, payload)
		},
	}
}

type postEvent struct {
	PostID string `json:"post_id"`
	BlogID string `json:"blog_id"`
}

func decodePostEvent(payload json.RawMessage) (postEvent, error) {
	var event postEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return postEvent{}, fmt.Errorf("invalid post event payload: %w", err)
	}
	if event.PostID == "" || event.BlogID == "" {
		return postEvent{}, errors.New("post event requires post_id and blog_id")
	}
	return event, nil
}

func handlePostPublished(ctx context.Context, deps Dependencies, payload json.RawMessage) error {
	event, err := decodePostEvent(payload)
	if err != nil {
		return err
	}
	if err := drainSearchJobs(ctx, deps, "POSTS", event.PostID); err != nil {
		return err
	}
	if deps.Treescan != nil {
		if err := deps.Treescan.MarkPost(ctx, event.PostID, false); err != nil {
			var fault *treescan.Fault
			if !errors.As(err, &fault) || fault.Code != "POST_NOT_FOUND" {
				return err
			}
		}
	}
	return fanoutPostNotifications(ctx, deps, event)
}

func handlePostUnpublished(ctx context.Context, deps Dependencies, payload json.RawMessage) error {
	event, err := decodePostEvent(payload)
	if err != nil {
		return err
	}
	if err := drainSearchJobs(ctx, deps, "POSTS", event.PostID); err != nil {
		return err
	}
	if deps.Treescan != nil {
		if err := deps.Treescan.MarkPost(ctx, event.PostID, true); err != nil {
			var fault *treescan.Fault
			if !errors.As(err, &fault) || fault.Code != "POST_NOT_FOUND" {
				return err
			}
		}
	}
	return nil
}

func handleMediaProcess(ctx context.Context, deps Dependencies, payload json.RawMessage) error {
	var event struct {
		AssetID string `json:"asset_id"`
	}
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("invalid media event payload: %w", err)
	}
	if event.AssetID == "" {
		return errors.New("media event requires asset_id")
	}
	if deps.Media == nil {
		return errors.New("media processor is not configured")
	}
	return deps.Media.ProcessAsset(ctx, event.AssetID)
}

// drainSearchJobs applies pending index jobs for one document. When
// OpenSearch is disabled the jobs stay PENDING (SQL search keeps serving)
// and the event succeeds so unrelated work is not blocked.
func drainSearchJobs(ctx context.Context, deps Dependencies, targetIndex, documentID string) error {
	rows, err := deps.Pool.Query(ctx, `
		SELECT id, operation
		FROM search_index_jobs
		WHERE target_index = $1 AND document_id = $2 AND status IN ('PENDING', 'FAILED')
		ORDER BY created_at ASC
		FOR UPDATE SKIP LOCKED
	`, targetIndex, documentID)
	if err != nil {
		return err
	}
	type job struct {
		id        string
		operation string
	}
	jobs := make([]job, 0)
	for rows.Next() {
		var j job
		if err := rows.Scan(&j.id, &j.operation); err != nil {
			rows.Close()
			return err
		}
		jobs = append(jobs, j)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	for _, j := range jobs {
		if !deps.OpenSearch.Enabled() {
			return nil
		}
		if err := applySearchJob(ctx, deps, targetIndex, documentID, j); err != nil {
			_, _ = deps.Pool.Exec(ctx, `UPDATE search_index_jobs
				SET status = 'FAILED', attempt_count = attempt_count + 1,
				    last_error = $2, processed_at = NOW()
				WHERE id = $1`, j.id, err.Error())
			return err
		}
		_, _ = deps.Pool.Exec(ctx, `UPDATE search_index_jobs
			SET status = 'COMPLETED', attempt_count = attempt_count + 1,
			    last_error = NULL, processed_at = NOW()
			WHERE id = $1`, j.id)
	}
	return nil
}

func applySearchJob(ctx context.Context, deps Dependencies, targetIndex, documentID string, j struct {
	id        string
	operation string
},
) error {
	switch targetIndex {
	case "POSTS":
		if j.operation == "DELETE" {
			return deps.OpenSearch.DeletePost(ctx, documentID)
		}
		doc, err := loadPostDocument(ctx, deps.Pool, documentID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				// The post vanished after the job was queued; make sure no
				// stale document stays behind and treat it as done.
				return deps.OpenSearch.DeletePost(ctx, documentID)
			}
			return err
		}
		return deps.OpenSearch.IndexPost(ctx, documentID, doc)
	case "BLOGS":
		if j.operation == "DELETE" {
			return deps.OpenSearch.DeleteBlog(ctx, documentID)
		}
		doc, err := loadBlogDocument(ctx, deps.Pool, documentID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return deps.OpenSearch.DeleteBlog(ctx, documentID)
			}
			return err
		}
		return deps.OpenSearch.IndexBlog(ctx, documentID, doc)
	default:
		return fmt.Errorf("unsupported search target %s", targetIndex)
	}
}

func loadPostDocument(ctx context.Context, pool *pgxpool.Pool, postID string) (opensearch.PostDocument, error) {
	var doc opensearch.PostDocument
	var category *string
	var publishedAt *string
	err := pool.QueryRow(ctx, `
		SELECT p.title, p.subtitle, p.excerpt, p.content_plain, p.blog_id, b.handle,
		       bc.name, p.visibility, p.published_at::text
		FROM posts p
		JOIN blogs b ON b.id = p.blog_id
		LEFT JOIN blog_categories bc ON bc.id = p.category_id
		WHERE p.id = $1
	`, postID).Scan(
		&doc.Title, &doc.Subtitle, &doc.Excerpt, &doc.Content, &doc.BlogID,
		&doc.BlogHandle, &category, &doc.Visibility, &publishedAt,
	)
	if err != nil {
		return opensearch.PostDocument{}, err
	}
	if category != nil {
		doc.Category = *category
	}
	if publishedAt != nil {
		doc.PublishedAt = *publishedAt
	}
	rows, err := pool.Query(ctx, `
		SELECT t.slug FROM post_tags pt JOIN tags t ON t.id = pt.tag_id WHERE pt.post_id = $1
	`, postID)
	if err == nil {
		defer rows.Close()
		doc.Tags = make([]string, 0)
		for rows.Next() {
			var slug string
			if err := rows.Scan(&slug); err == nil {
				doc.Tags = append(doc.Tags, slug)
			}
		}
	}
	return doc, nil
}

func loadBlogDocument(ctx context.Context, pool *pgxpool.Pool, blogID string) (opensearch.BlogDocument, error) {
	var doc opensearch.BlogDocument
	err := pool.QueryRow(ctx, `SELECT title, handle, description FROM blogs WHERE id = $1`, blogID).
		Scan(&doc.Title, &doc.Handle, &doc.Description)
	return doc, err
}

// fanoutPostNotifications inserts one SUBSCRIPTION notification per active
// subscriber. The NOT EXISTS guard keeps at-least-once delivery idempotent.
func fanoutPostNotifications(ctx context.Context, deps Dependencies, event postEvent) error {
	var title, slug, handle string
	var excerpt string
	var authorID string
	err := deps.Pool.QueryRow(ctx, `
		SELECT p.title, p.slug, b.handle, p.excerpt, p.author_user_id
		FROM posts p JOIN blogs b ON b.id = p.blog_id
		WHERE p.id = $1 AND p.blog_id = $2
	`, event.PostID, event.BlogID).Scan(&title, &slug, &handle, &excerpt, &authorID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}
	base := strings.TrimRight(deps.AppBaseURL, "/")
	if base == "" {
		base = "http://localhost:3000"
	}
	actionURL := fmt.Sprintf("%s/b/%s/%s", base, handle, slug)
	body := excerpt
	if body == "" {
		body = "새 글이 발행되었습니다."
	}
	_, err = deps.Pool.Exec(ctx, `
		INSERT INTO notifications (recipient_user_id, category, group_key, dedup_key, title, body, action_url)
		SELECT s.subscriber_user_id, 'SUBSCRIPTION', 'blog:' || $1, 'post-published:' || $2, $3, $4, $5
		FROM subscriptions s
		WHERE s.target_blog_id = $1 AND s.status = 'ACTIVE' AND s.notify_in_app = TRUE
		  AND s.subscriber_user_id <> $6
		  AND NOT EXISTS (
			SELECT 1 FROM notifications n
			WHERE n.recipient_user_id = s.subscriber_user_id AND n.dedup_key = 'post-published:' || $2
		  )
		LIMIT 5000
	`, event.BlogID, event.PostID, title, body, actionURL, authorID)
	return err
}
