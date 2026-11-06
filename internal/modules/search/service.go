package search

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"treeblog/internal/platform/opensearch"
)

type Fault struct {
	Status  int
	Code    string
	Message string
}

func (f *Fault) Error() string {
	return f.Message
}

type SearchPostItem struct {
	ID             string     `json:"id"`
	BlogID         string     `json:"blog_id"`
	BlogTitle      string     `json:"blog_title"`
	BlogHandle     string     `json:"blog_handle"`
	AuthorName     string     `json:"author_name"`
	Slug           string     `json:"slug"`
	Title          string     `json:"title"`
	Subtitle       string     `json:"subtitle"`
	Excerpt        string     `json:"excerpt"`
	CoverImageURL  *string    `json:"cover_image_url,omitempty"`
	CategoryName   *string    `json:"category_name,omitempty"`
	Tags           []string   `json:"tags"`
	WordCount      int        `json:"word_count"`
	ReadingTimeSec int        `json:"reading_time_sec"`
	ViewCount      int64      `json:"view_count"`
	ReactionCount  int        `json:"reaction_count"`
	CommentCount   int        `json:"comment_count"`
	PublishedAt    *time.Time `json:"published_at,omitempty"`
	Score          float64    `json:"score,omitempty"`
}

type SearchBlogItem struct {
	ID          string `json:"id"`
	Handle      string `json:"handle"`
	Title       string `json:"title"`
	Description string `json:"description"`
	PostCount   int64  `json:"post_count"`
	AvatarURL   string `json:"avatar_url,omitempty"`
}

type SearchTagItem struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	PostCount int64  `json:"post_count"`
}

type ClickEventInput struct {
	QueryText    string `json:"query_text"`
	ClickedPostID string `json:"clicked_post_id"`
	RankPosition int    `json:"rank_position"`
	SessionHash  string `json:"session_hash"`
	DwellTimeSec int    `json:"dwell_time_sec"`
}

type Service struct {
	pool        *pgxpool.Pool
	opensearch  *opensearch.Client
}

// NewService builds a search service. A nil OpenSearch client disables
// ranked retrieval and every query uses the SQL fallback.
func NewService(pool *pgxpool.Pool, osClient *opensearch.Client) *Service {
	return &Service{pool: pool, opensearch: osClient}
}

// ComputeSimHash64 computes a 64-bit SimHash of text for near-duplicate content detection.
func ComputeSimHash64(text string) uint64 {
	words := strings.Fields(strings.ToLower(text))
	if len(words) == 0 {
		return 0
	}

	var v [64]int
	for _, word := range words {
		h := md5.Sum([]byte(word))
		hash64 := binary.BigEndian.Uint64(h[:8])

		for i := 0; i < 64; i++ {
			bit := (hash64 >> i) & 1
			if bit == 1 {
				v[i]++
			} else {
				v[i]--
			}
		}
	}

	var simhash uint64
	for i := 0; i < 64; i++ {
		if v[i] > 0 {
			simhash |= (1 << i)
		}
	}
	return simhash
}

// HammingDistance calculates the Hamming distance between two 64-bit SimHashes.
func HammingDistance(a, b uint64) int {
	xor := a ^ b
	dist := 0
	for xor != 0 {
		dist += int(xor & 1)
		xor >>= 1
	}
	return dist
}

func (s *Service) SearchPosts(ctx context.Context, query, blogScopeID string, limit int, offset int) ([]SearchPostItem, int, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return []SearchPostItem{}, 0, nil
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	if s.pool == nil {
		return nil, 0, &Fault{Status: http.StatusServiceUnavailable, Code: "SERVICE_UNAVAILABLE", Message: "Database is unavailable"}
	}

	// Update search_queries count asynchronously
	go func(q string) {
		_, _ = s.pool.Exec(context.Background(), `
			INSERT INTO search_queries (normalized_query, search_count, last_searched_at)
			VALUES ($1, 1, NOW())
			ON CONFLICT (normalized_query, language_code) DO UPDATE
			SET search_count = search_queries.search_count + 1, last_searched_at = NOW()
		`, strings.ToLower(q))
	}(query)

	if s.opensearch.Enabled() {
		if ranked, total, err := s.searchPostsRanked(ctx, query, blogScopeID, limit, offset); err == nil {
			return ranked, total, nil
		}
		// Ranked retrieval failed; fall through to the SQL fallback so
		// search stays available during an OpenSearch outage.
	}
	return s.searchPostsSQL(ctx, query, blogScopeID, limit, offset)
}

// searchPostsRanked resolves ranked IDs from OpenSearch and hydrates them
// from PostgreSQL, preserving rank order.
func (s *Service) searchPostsRanked(ctx context.Context, query, blogScopeID string, limit, offset int) ([]SearchPostItem, int, error) {
	hits, total, err := s.opensearch.SearchPosts(ctx, query, blogScopeID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	if len(hits) == 0 {
		return []SearchPostItem{}, total, nil
	}
	ids := make([]string, 0, len(hits))
	scores := make(map[string]float64, len(hits))
	for _, hit := range hits {
		ids = append(ids, hit.ID)
		scores[hit.ID] = hit.Score
	}
	rows, err := s.pool.Query(ctx, `
		SELECT p.id, p.blog_id, b.title, b.handle, COALESCE(up.display_name, u.email::text), p.slug, p.title, p.subtitle,
		       p.excerpt, bc.name, p.word_count, p.reading_time_sec, p.view_count,
		       p.reaction_count, p.comment_count, p.published_at, p.cover_media_id
		FROM posts p
		JOIN blogs b ON b.id = p.blog_id
		JOIN users u ON u.id = p.author_user_id
		LEFT JOIN user_profiles up ON up.user_id = u.id
		LEFT JOIN blog_categories bc ON bc.id = p.category_id
		WHERE p.id = ANY($1) AND p.status = 'PUBLISHED' AND p.visibility = 'PUBLIC' AND p.trashed_at IS NULL
		ORDER BY array_position($1::uuid[], p.id)
	`, ids)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	results, err := scanPostItems(rows)
	if err != nil {
		return nil, 0, err
	}
	for i := range results {
		results[i].Score = scores[results[i].ID]
	}
	if err := s.fillTags(ctx, results); err != nil {
		return nil, 0, err
	}
	return results, total, nil
}

func (s *Service) searchPostsSQL(ctx context.Context, query, blogScopeID string, limit, offset int) ([]SearchPostItem, int, error) {
	sql := `
		SELECT p.id, p.blog_id, b.title, b.handle, COALESCE(up.display_name, u.email::text), p.slug, p.title, p.subtitle,
		       p.excerpt, bc.name, p.word_count, p.reading_time_sec, p.view_count,
		       p.reaction_count, p.comment_count, p.published_at, p.cover_media_id
		FROM posts p
		JOIN blogs b ON b.id = p.blog_id
		JOIN users u ON u.id = p.author_user_id
		LEFT JOIN user_profiles up ON up.user_id = u.id
		LEFT JOIN blog_categories bc ON bc.id = p.category_id
		WHERE p.status = 'PUBLISHED' AND p.visibility = 'PUBLIC' AND p.trashed_at IS NULL
		  AND (p.title ILIKE '%' || $1 || '%' OR p.content_plain ILIKE '%' || $1 || '%' OR p.subtitle ILIKE '%' || $1 || '%')
	`
	args := []any{query}
	argIdx := 2

	if blogScopeID != "" {
		sql += fmt.Sprintf(" AND p.blog_id = $%d", argIdx)
		args = append(args, blogScopeID)
		argIdx++
	}

	sql += fmt.Sprintf(" ORDER BY p.published_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	results, err := scanPostItems(rows)
	if err != nil {
		return nil, 0, err
	}
	if err := s.fillTags(ctx, results); err != nil {
		return nil, 0, err
	}
	return results, len(results), nil
}

func scanPostItems(rows pgx.Rows) ([]SearchPostItem, error) {
	results := make([]SearchPostItem, 0)
	for rows.Next() {
		var item SearchPostItem
		var catName *string
		var coverMediaID *string
		if err := rows.Scan(
			&item.ID, &item.BlogID, &item.BlogTitle, &item.BlogHandle, &item.AuthorName,
			&item.Slug, &item.Title, &item.Subtitle, &item.Excerpt, &catName,
			&item.WordCount, &item.ReadingTimeSec, &item.ViewCount,
			&item.ReactionCount, &item.CommentCount, &item.PublishedAt, &coverMediaID,
		); err != nil {
			return nil, err
		}
		item.CategoryName = catName
		item.CoverImageURL = coverImageURL(coverMediaID)
		results = append(results, item)
	}
	return results, rows.Err()
}

// coverImageURL resolves a cover media id to its public delivery URL.
func coverImageURL(coverMediaID *string) *string {
	if coverMediaID == nil || *coverMediaID == "" {
		return nil
	}
	url := "/api/v1/public/media/" + *coverMediaID
	return &url
}

// fillTags hydrates tag names for listed posts in one query.
func (s *Service) fillTags(ctx context.Context, items []SearchPostItem) error {
	if len(items) == 0 {
		return nil
	}
	if s.pool == nil {
		return &Fault{Status: http.StatusServiceUnavailable, Code: "SERVICE_UNAVAILABLE", Message: "Database is unavailable"}
	}
	ids := make([]string, 0, len(items))
	byID := make(map[string]*SearchPostItem, len(items))
	for i := range items {
		ids = append(ids, items[i].ID)
		byID[items[i].ID] = &items[i]
		items[i].Tags = []string{}
	}
	rows, err := s.pool.Query(ctx, `
		SELECT pt.post_id, t.name
		FROM post_tags pt JOIN tags t ON t.id = pt.tag_id
		WHERE pt.post_id = ANY($1)
		ORDER BY pt.sort_order ASC
	`, ids)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var postID, name string
		if err := rows.Scan(&postID, &name); err != nil {
			return err
		}
		if item, ok := byID[postID]; ok {
			item.Tags = append(item.Tags, name)
		}
	}
	return rows.Err()
}

func (s *Service) SearchBlogs(ctx context.Context, query string, limit int) ([]SearchBlogItem, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return []SearchBlogItem{}, nil
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	if s.pool == nil {
		return nil, &Fault{Status: http.StatusServiceUnavailable, Code: "SERVICE_UNAVAILABLE", Message: "Database is unavailable"}
	}

	rows, err := s.pool.Query(ctx, `
		SELECT id, handle, title, description, post_count
		FROM blogs
		WHERE status = 'ACTIVE'
		  AND (title ILIKE '%' || $1 || '%' OR handle ILIKE '%' || $1 || '%' OR description ILIKE '%' || $1 || '%')
		ORDER BY post_count DESC
		LIMIT $2
	`, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	blogs := make([]SearchBlogItem, 0)
	for rows.Next() {
		var b SearchBlogItem
		if err := rows.Scan(&b.ID, &b.Handle, &b.Title, &b.Description, &b.PostCount); err != nil {
			return nil, err
		}
		blogs = append(blogs, b)
	}
	return blogs, rows.Err()
}

func (s *Service) SearchTags(ctx context.Context, query string, limit int) ([]SearchTagItem, error) {
	query = strings.TrimSpace(query)
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	if s.pool == nil {
		return nil, &Fault{Status: http.StatusServiceUnavailable, Code: "SERVICE_UNAVAILABLE", Message: "Database is unavailable"}
	}

	sql := `SELECT id, name, slug, post_count FROM tags WHERE 1=1`
	args := []any{}
	if query != "" {
		sql += ` AND (name ILIKE '%' || $1 || '%' OR slug ILIKE '%' || $1 || '%')`
		args = append(args, query)
		sql += fmt.Sprintf(` ORDER BY post_count DESC LIMIT $%d`, len(args)+1)
		args = append(args, limit)
	} else {
		sql += fmt.Sprintf(` ORDER BY post_count DESC LIMIT $1`)
		args = append(args, limit)
	}

	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tags := make([]SearchTagItem, 0)
	for rows.Next() {
		var t SearchTagItem
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug, &t.PostCount); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}

func (s *Service) Suggest(ctx context.Context, prefix string, limit int) ([]string, error) {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		return []string{}, nil
	}
	if limit <= 0 || limit > 10 {
		limit = 5
	}

	if s.pool == nil {
		return nil, &Fault{Status: http.StatusServiceUnavailable, Code: "SERVICE_UNAVAILABLE", Message: "Database is unavailable"}
	}

	rows, err := s.pool.Query(ctx, `
		SELECT normalized_query FROM search_queries
		WHERE normalized_query ILIKE $1 || '%' AND is_blocked_spam_term = FALSE
		ORDER BY search_count DESC
		LIMIT $2
	`, prefix, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	suggestions := make([]string, 0)
	for rows.Next() {
		var q string
		if err := rows.Scan(&q); err == nil {
			suggestions = append(suggestions, q)
		}
	}
	return suggestions, rows.Err()
}

func (s *Service) LogClickEvent(ctx context.Context, input ClickEventInput) error {
	if s.pool == nil {
		return &Fault{Status: http.StatusServiceUnavailable, Code: "SERVICE_UNAVAILABLE", Message: "Database is unavailable"}
	}
	if input.QueryText == "" || input.ClickedPostID == "" || input.RankPosition <= 0 {
		return &Fault{Status: http.StatusBadRequest, Code: "INVALID_CLICK_EVENT", Message: "Missing required click event fields"}
	}

	sessionHash := input.SessionHash
	if len(sessionHash) != 64 {
		h := md5.Sum([]byte(fmt.Sprintf("%s-%d", input.ClickedPostID, time.Now().UnixNano())))
		sessionHash = fmt.Sprintf("%032x%032x", h, h)
	}

	_, err := s.pool.Exec(ctx, `
		INSERT INTO search_click_events (query_text, clicked_post_id, rank_position, session_hash, dwell_time_sec)
		VALUES ($1, $2, $3, $4, $5)
	`, input.QueryText, input.ClickedPostID, input.RankPosition, sessionHash, input.DwellTimeSec)
	return err
}

func (s *Service) GetTrendingPosts(ctx context.Context, limit int) ([]SearchPostItem, error) {
	if limit <= 0 || limit > 30 {
		limit = 10
	}
	if s.pool == nil {
		return nil, &Fault{Status: http.StatusServiceUnavailable, Code: "SERVICE_UNAVAILABLE", Message: "Database is unavailable"}
	}

	rows, err := s.pool.Query(ctx, `
		SELECT p.id, p.blog_id, b.title, b.handle, COALESCE(up.display_name, u.email::text), p.slug, p.title, p.subtitle,
		       p.excerpt, bc.name, p.word_count, p.reading_time_sec, p.view_count,
		       p.reaction_count, p.comment_count, p.published_at, p.cover_media_id
		FROM posts p
		JOIN blogs b ON b.id = p.blog_id
		JOIN users u ON u.id = p.author_user_id
		LEFT JOIN user_profiles up ON up.user_id = u.id
		LEFT JOIN blog_categories bc ON bc.id = p.category_id
		WHERE p.status = 'PUBLISHED' AND p.visibility = 'PUBLIC' AND p.trashed_at IS NULL
		  AND p.published_at >= NOW() - INTERVAL '7 days'
		ORDER BY (p.view_count + p.reaction_count * 5 + p.comment_count * 10) DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list, err := scanPostItems(rows)
	if err != nil {
		return nil, err
	}
	if err := s.fillTags(ctx, list); err != nil {
		return nil, err
	}
	return list, nil
}

func (s *Service) GetPopularBlogs(ctx context.Context, limit int) ([]SearchBlogItem, error) {
	if limit <= 0 || limit > 30 {
		limit = 10
	}
	if s.pool == nil {
		return nil, &Fault{Status: http.StatusServiceUnavailable, Code: "SERVICE_UNAVAILABLE", Message: "Database is unavailable"}
	}

	rows, err := s.pool.Query(ctx, `
		SELECT id, handle, title, description, post_count
		FROM blogs
		WHERE status = 'ACTIVE'
		ORDER BY post_count DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]SearchBlogItem, 0)
	for rows.Next() {
		var b SearchBlogItem
		if err := rows.Scan(&b.ID, &b.Handle, &b.Title, &b.Description, &b.PostCount); err != nil {
			return nil, err
		}
		list = append(list, b)
	}
	return list, rows.Err()
}

func (s *Service) GetRelatedPosts(ctx context.Context, postID string, limit int) ([]SearchPostItem, error) {
	if limit <= 0 || limit > 10 {
		limit = 5
	}
	if s.pool == nil {
		return nil, &Fault{Status: http.StatusServiceUnavailable, Code: "SERVICE_UNAVAILABLE", Message: "Database is unavailable"}
	}

	var blogID string
	var categoryID *string
	err := s.pool.QueryRow(ctx, `SELECT blog_id, category_id FROM posts WHERE id = $1`, postID).Scan(&blogID, &categoryID)
	if err != nil {
		return nil, &Fault{Status: http.StatusNotFound, Code: "POST_NOT_FOUND", Message: "Post not found"}
	}

	rows, err := s.pool.Query(ctx, `
		SELECT p.id, p.blog_id, b.title, b.handle, COALESCE(up.display_name, u.email::text), p.slug, p.title, p.subtitle,
		       p.excerpt, bc.name, p.word_count, p.reading_time_sec, p.view_count,
		       p.reaction_count, p.comment_count, p.published_at, p.cover_media_id
		FROM posts p
		JOIN blogs b ON b.id = p.blog_id
		JOIN users u ON u.id = p.author_user_id
		LEFT JOIN user_profiles up ON up.user_id = u.id
		LEFT JOIN blog_categories bc ON bc.id = p.category_id
		WHERE p.blog_id = $1 AND p.id <> $2 AND p.status = 'PUBLISHED' AND p.visibility = 'PUBLIC' AND p.trashed_at IS NULL
		  AND ($3::uuid IS NULL OR p.category_id = $3)
		ORDER BY p.published_at DESC
		LIMIT $4
	`, blogID, postID, categoryID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list, err := scanPostItems(rows)
	if err != nil {
		return nil, err
	}
	if err := s.fillTags(ctx, list); err != nil {
		return nil, err
	}
	return list, nil
}
