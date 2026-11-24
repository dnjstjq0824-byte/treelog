package seo

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// PostMeta is the crawler-facing metadata for one post.
type PostMeta struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Canonical   string   `json:"canonical_url"`
	ImageURL    *string  `json:"image_url,omitempty"`
	AuthorName  string   `json:"author_name"`
	PublishedAt *string  `json:"published_at,omitempty"`
	UpdatedAt   string   `json:"updated_at"`
	Tags        []string `json:"tags"`
	// Robots is "index,follow" for public posts and "noindex,nofollow" for
	// every restricted visibility.
	Robots string `json:"robots"`
	// JSONLD carries BlogPosting + BreadcrumbList blocks for <script> injection.
	JSONLD []any `json:"json_ld"`
}

// PostMeta resolves crawler metadata for a published post. Private posts
// 404 (existence is not confirmed to crawlers); other restricted
// visibilities return metadata pinned to noindex.
func (s *Service) PostMeta(ctx context.Context, blogKey, slug string) (*PostMeta, error) {
	if err := s.requirePool(); err != nil {
		return nil, err
	}
	var (
		postID, handle, title, excerpt, visibility, authorName string
		seoTitle, seoDesc                                       *string
		coverMediaID                                            *string
		publishedAt                                             *time.Time
		updatedAt                                               time.Time
	)
	err := s.pool.QueryRow(ctx, `
		SELECT p.id, b.handle, p.title, p.excerpt, p.visibility,
		       COALESCE(up.display_name, u.email::text),
		       p.seo_title, p.seo_description, p.cover_media_id,
		       p.published_at, p.updated_at
		FROM posts p
		JOIN blogs b ON b.id = p.blog_id
		JOIN users u ON u.id = p.author_user_id
		LEFT JOIN user_profiles up ON up.user_id = u.id
		WHERE (b.id::text = $1 OR b.handle = $1) AND p.slug = $2
		  AND p.status = 'PUBLISHED' AND p.trashed_at IS NULL
	`, blogKey, slug).Scan(
		&postID, &handle, &title, &excerpt, &visibility, &authorName,
		&seoTitle, &seoDesc, &coverMediaID, &publishedAt, &updatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, &Fault{Status: http.StatusNotFound, Code: "POST_NOT_FOUND", Message: "Post not found or unpublished"}
	}
	if err != nil {
		return nil, err
	}
	if visibility == "PRIVATE" {
		return nil, &Fault{Status: http.StatusNotFound, Code: "POST_NOT_FOUND", Message: "Post not found or unpublished"}
	}
	metaTitle := title
	if seoTitle != nil && strings.TrimSpace(*seoTitle) != "" {
		metaTitle = *seoTitle
	}
	metaDesc := excerpt
	if seoDesc != nil && strings.TrimSpace(*seoDesc) != "" {
		metaDesc = *seoDesc
	}
	canonical := s.appBase + "/b/" + handle + "/" + slug
	var imageURL *string
	if coverMediaID != nil && *coverMediaID != "" {
		cover := s.appBase + "/api/v1/public/media/" + *coverMediaID + "?variant=W1200"
		imageURL = &cover
	}
	tags := make([]string, 0)
	rows, err := s.pool.Query(ctx, `
		SELECT t.name FROM tags t JOIN post_tags pt ON pt.tag_id = t.id
		WHERE pt.post_id = $1 ORDER BY pt.sort_order ASC`, postID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err == nil {
				tags = append(tags, name)
			}
		}
	}
	robots := "index,follow"
	if visibility != "PUBLIC" {
		robots = "noindex,nofollow"
	}
	var published *string
	if publishedAt != nil {
		formatted := publishedAt.UTC().Format(time.RFC3339)
		published = &formatted
	}
	meta := &PostMeta{
		Title:       metaTitle,
		Description: metaDesc,
		Canonical:   canonical,
		ImageURL:    imageURL,
		AuthorName:  authorName,
		PublishedAt: published,
		UpdatedAt:   updatedAt.UTC().Format(time.RFC3339),
		Tags:        tags,
		Robots:      robots,
	}
	meta.JSONLD = buildPostJSONLD(s.appBase, handle, title, metaDesc, canonical, imageURL, authorName, publishedAt, updatedAt, tags)
	return meta, nil
}

// buildPostJSONLD renders BlogPosting + BreadcrumbList blocks. It is pure so
// the schema stays pinned by unit tests without a database.
func buildPostJSONLD(appBase, handle, title, description, canonical string, imageURL *string, authorName string, publishedAt *time.Time, updatedAt time.Time, tags []string) []any {
	blogHome := appBase + "/b/" + handle
	posting := map[string]any{
		"@context":      "https://schema.org",
		"@type":         "BlogPosting",
		"headline":      title,
		"description":   description,
		"url":           canonical,
		"mainEntityOfPage": map[string]any{"@type": "WebPage", "@id": canonical},
		"author":        map[string]any{"@type": "Person", "name": authorName},
		"dateModified":  updatedAt.UTC().Format(time.RFC3339),
	}
	if imageURL != nil {
		posting["image"] = *imageURL
	}
	if publishedAt != nil {
		posting["datePublished"] = publishedAt.UTC().Format(time.RFC3339)
	}
	if len(tags) > 0 {
		keywords := make([]string, 0, len(tags))
		keywords = append(keywords, tags...)
		posting["keywords"] = strings.Join(keywords, ", ")
	}
	breadcrumb := map[string]any{
		"@context": "https://schema.org",
		"@type":    "BreadcrumbList",
		"itemListElement": []any{
			map[string]any{"@type": "ListItem", "position": 1, "name": "TreeBlog", "item": appBase + "/"},
			map[string]any{"@type": "ListItem", "position": 2, "name": handle, "item": blogHome},
			map[string]any{"@type": "ListItem", "position": 3, "name": title, "item": canonical},
		},
	}
	return []any{posting, breadcrumb}
}
