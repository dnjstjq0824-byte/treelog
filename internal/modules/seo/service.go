package seo

import (
	"context"
	"encoding/xml"
	"errors"
	"net/http"
	"strings"
	"time"

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

// Service generates sitemaps, robots directives, and link metadata.
type Service struct {
	pool    *pgxpool.Pool
	appBase string
}

// NewService builds an SEO service. appBase is the public user-web origin.
func NewService(pool *pgxpool.Pool, appBase string) (*Service, error) {
	if pool == nil {
		return nil, errors.New("seo service requires a database pool")
	}
	appBase = strings.TrimRight(strings.TrimSpace(appBase), "/")
	if appBase == "" {
		appBase = "http://localhost:3000"
	}
	return &Service{pool: pool, appBase: appBase}, nil
}

func (s *Service) requirePool() error {
	if s.pool == nil {
		return &Fault{Status: http.StatusServiceUnavailable, Code: "SERVICE_UNAVAILABLE", Message: "Database is unavailable"}
	}
	return nil
}

type urlSet struct {
	XMLName xml.Name `xml:"urlset"`
	Xmlns   string   `xml:"xmlns,attr"`
	URLs    []sitemapURL `xml:"url"`
}

type sitemapURL struct {
	Loc        string `xml:"loc"`
	LastMod    string `xml:"lastmod,omitempty"`
	ChangeFreq string `xml:"changefreq,omitempty"`
	Priority   string `xml:"priority,omitempty"`
}

// Sitemap renders the public sitemap covering active blogs and the newest
// public posts.
func (s *Service) Sitemap(ctx context.Context) (string, error) {
	if err := s.requirePool(); err != nil {
		return "", err
	}
	set := urlSet{Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9"}
	set.URLs = append(set.URLs, sitemapURL{Loc: s.appBase + "/", ChangeFreq: "hourly", Priority: "1.0"})

	blogRows, err := s.pool.Query(ctx, `
		SELECT handle, updated_at FROM blogs WHERE status = 'ACTIVE' ORDER BY updated_at DESC LIMIT 5000
	`)
	if err != nil {
		return "", err
	}
	for blogRows.Next() {
		var handle string
		var updated time.Time
		if err := blogRows.Scan(&handle, &updated); err != nil {
			blogRows.Close()
			return "", err
		}
		set.URLs = append(set.URLs, sitemapURL{
			Loc: s.appBase + "/b/" + handle,
			LastMod: updated.UTC().Format("2006-01-02"),
			ChangeFreq: "daily", Priority: "0.8",
		})
	}
	blogRows.Close()
	if err := blogRows.Err(); err != nil {
		return "", err
	}

	postRows, err := s.pool.Query(ctx, `
		SELECT b.handle, p.slug, p.updated_at
		FROM posts p JOIN blogs b ON b.id = p.blog_id
		WHERE p.status = 'PUBLISHED' AND p.visibility = 'PUBLIC' AND p.trashed_at IS NULL
		  AND b.status = 'ACTIVE'
		ORDER BY p.published_at DESC NULLS LAST LIMIT 5000
	`)
	if err != nil {
		return "", err
	}
	for postRows.Next() {
		var handle, slug string
		var updated time.Time
		if err := postRows.Scan(&handle, &slug, &updated); err != nil {
			postRows.Close()
			return "", err
		}
		set.URLs = append(set.URLs, sitemapURL{
			Loc: s.appBase + "/b/" + handle + "/" + slug,
			LastMod: updated.UTC().Format("2006-01-02"),
			ChangeFreq: "weekly", Priority: "0.6",
		})
	}
	postRows.Close()
	if err := postRows.Err(); err != nil {
		return "", err
	}

	encoded, err := xml.MarshalIndent(set, "", "  ")
	if err != nil {
		return "", err
	}
	return xml.Header + string(encoded), nil
}

// Robots renders the robots.txt body.
func (s *Service) Robots() string {
	return "User-agent: *\nAllow: /\nDisallow: /api/\nSitemap: " + s.appBase + "/sitemap.xml\n"
}

// BlogMeta is the link-preview metadata for a blog home.
type BlogMeta struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Canonical   string `json:"canonical_url"`
	ImageURL    *string `json:"image_url,omitempty"`
	Handle      string `json:"handle"`
}

// BlogMeta returns metadata for a blog home page.
func (s *Service) BlogMeta(ctx context.Context, blogKey string) (*BlogMeta, error) {
	if err := s.requirePool(); err != nil {
		return nil, err
	}
	var meta BlogMeta
	var description string
	err := s.pool.QueryRow(ctx, `
		SELECT handle, title, description FROM blogs
		WHERE id::text = $1 OR handle = $1
	`, blogKey).Scan(&meta.Handle, &meta.Title, &description)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, &Fault{Status: http.StatusNotFound, Code: "BLOG_NOT_FOUND", Message: "Blog not found"}
	}
	if err != nil {
		return nil, err
	}
	meta.Description = description
	if meta.Description == "" {
		meta.Description = meta.Title + " - TreeBlog에서 만나요"
	}
	meta.Canonical = s.appBase + "/b/" + meta.Handle
	return &meta, nil
}
