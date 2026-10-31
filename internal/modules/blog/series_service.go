package blog

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type Series struct {
	ID          string    `json:"id"`
	BlogID      string    `json:"blog_id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	CoverMediaID *string  `json:"cover_media_id"`
	IsCompleted bool      `json:"is_completed"`
	PostCount   int       `json:"post_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TagInfo struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	PostCount int64     `json:"post_count"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *Service) ListSeries(ctx context.Context, blogID string) ([]Series, error) {
	rows, err := s.db.Query(ctx, `SELECT id, blog_id, title, slug, description,
		cover_media_id, is_completed, post_count, created_at, updated_at
		FROM series
		WHERE blog_id = $1
		ORDER BY created_at DESC`, blogID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Series
	for rows.Next() {
		var item Series
		if err := rows.Scan(
			&item.ID, &item.BlogID, &item.Title, &item.Slug, &item.Description,
			&item.CoverMediaID, &item.IsCompleted, &item.PostCount, &item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return list, nil
}

func (s *Service) CreateSeries(ctx context.Context, blogID, title, slug, description string) (*Series, error) {
	title = strings.TrimSpace(title)
	if len(title) < 1 || len(title) > 160 {
		return nil, &Fault{Status: 400, Code: "INVALID_REQUEST_PAYLOAD", Message: "Series title must be between 1 and 160 characters"}
	}
	slug = strings.ToLower(strings.TrimSpace(slug))
	if len(slug) < 1 || len(slug) > 180 {
		return nil, &Fault{Status: 400, Code: "INVALID_REQUEST_PAYLOAD", Message: "Series slug must be between 1 and 180 characters"}
	}

	var item Series
	item.ID = uuid.NewString()
	item.BlogID = blogID
	item.Title = title
	item.Slug = slug
	item.Description = strings.TrimSpace(description)
	item.IsCompleted = false
	item.PostCount = 0

	err := s.db.QueryRow(ctx, `INSERT INTO series (
		id, blog_id, title, slug, description, is_completed, post_count
	) VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING created_at, updated_at`,
		item.ID, blogID, title, slug, item.Description, item.IsCompleted, item.PostCount,
	).Scan(&item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, &Fault{Status: 409, Code: "DUPLICATE_SERIES_SLUG", Message: "Series with this slug already exists"}
		}
		return nil, err
	}
	return &item, nil
}

func (s *Service) ListPopularTags(ctx context.Context, limit int) ([]TagInfo, error) {
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	rows, err := s.db.Query(ctx, `SELECT id, name, slug, post_count, created_at
		FROM tags
		WHERE post_count > 0
		ORDER BY post_count DESC, name ASC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []TagInfo
	for rows.Next() {
		var t TagInfo
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug, &t.PostCount, &t.CreatedAt); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, nil
}
