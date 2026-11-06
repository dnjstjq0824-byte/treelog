// Package opensearch is a minimal OpenSearch client using only the standard
// library. It manages the posts/blogs indices (Korean nori analysis) and
// ranked post retrieval; the search service falls back to SQL ILIKE when
// OpenSearch is unreachable.
package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Config describes the OpenSearch cluster endpoint.
type Config struct {
	// Address is the cluster base URL, e.g. http://localhost:9200.
	Address string
	// IndexPrefix namespaces indices, e.g. "treeblog".
	IndexPrefix string
	// Username/Password enable basic auth when the security plugin is on.
	Username string
	Password string
}

// Client executes OpenSearch requests.
type Client struct {
	cfg  Config
	http *http.Client
}

// New creates a Client. An empty address means OpenSearch is disabled and
// callers must use the SQL fallback.
func New(cfg Config) *Client {
	if cfg.IndexPrefix == "" {
		cfg.IndexPrefix = "treeblog"
	}
	return &Client{
		cfg:  cfg,
		http: &http.Client{Timeout: 10 * time.Second},
	}
}

// Enabled reports whether OpenSearch calls should be attempted.
func (c *Client) Enabled() bool {
	return c != nil && c.cfg.Address != ""
}

func (c *Client) base() string {
	return strings.TrimRight(c.cfg.Address, "/")
}

func (c *Client) postsIndex() string {
	return c.cfg.IndexPrefix + "-posts"
}

func (c *Client) blogsIndex() string {
	return c.cfg.IndexPrefix + "-blogs"
}

func (c *Client) do(ctx context.Context, method, path string, body any) (int, []byte, error) {
	if !c.Enabled() {
		return 0, nil, errors.New("opensearch is not configured")
	}
	if err := ctx.Err(); err != nil {
		return 0, nil, err
	}
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return 0, nil, err
		}
		reader = bytes.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.base()+path, reader)
	if err != nil {
		return 0, nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.cfg.Username != "" {
		req.SetBasicAuth(c.cfg.Username, c.cfg.Password)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("opensearch %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()
	payload, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return resp.StatusCode, nil, err
	}
	return resp.StatusCode, payload, nil
}

// EnsureIndices creates the posts/blogs indices with Korean analysis when
// they do not exist yet.
func (c *Client) EnsureIndices(ctx context.Context) error {
	for _, index := range []struct {
		name string
		body map[string]any
	}{
		{c.postsIndex(), postsIndexBody()},
		{c.blogsIndex(), blogsIndexBody()},
	} {
		status, _, err := c.do(ctx, http.MethodHead, "/"+index.name, nil)
		if err != nil {
			return err
		}
		if status == http.StatusOK {
			continue
		}
		status, payload, err := c.do(ctx, http.MethodPut, "/"+index.name, index.body)
		if err != nil {
			return err
		}
		if status != http.StatusOK && status != http.StatusCreated {
			return fmt.Errorf("create index %s: status %d: %.200s", index.name, status, payload)
		}
	}
	return nil
}

func koreanAnalyzer() map[string]any {
	return map[string]any{
		"analysis": map[string]any{
			"analyzer": map[string]any{
				"treeblog_korean": map[string]any{
					"type":      "custom",
					"tokenizer": "nori_tokenizer",
					"filter":    []string{"nori_part_of_speech", "lowercase"},
				},
			},
		},
	}
}

func postsIndexBody() map[string]any {
	return map[string]any{
		"settings": koreanAnalyzer(),
		"mappings": map[string]any{
			"properties": map[string]any{
				"title":        map[string]any{"type": "text", "analyzer": "treeblog_korean"},
				"subtitle":     map[string]any{"type": "text", "analyzer": "treeblog_korean"},
				"excerpt":      map[string]any{"type": "text", "analyzer": "treeblog_korean"},
				"content_plain": map[string]any{"type": "text", "analyzer": "treeblog_korean"},
				"blog_id":      map[string]any{"type": "keyword"},
				"blog_handle":  map[string]any{"type": "keyword"},
				"category":     map[string]any{"type": "keyword"},
				"tags":         map[string]any{"type": "keyword"},
				"visibility":   map[string]any{"type": "keyword"},
				"published_at": map[string]any{"type": "date"},
			},
		},
	}
}

func blogsIndexBody() map[string]any {
	return map[string]any{
		"settings": koreanAnalyzer(),
		"mappings": map[string]any{
			"properties": map[string]any{
				"title":       map[string]any{"type": "text", "analyzer": "treeblog_korean"},
				"handle":      map[string]any{"type": "keyword"},
				"description": map[string]any{"type": "text", "analyzer": "treeblog_korean"},
			},
		},
	}
}

// PostDocument is the indexed projection of a published post.
type PostDocument struct {
	Title       string   `json:"title"`
	Subtitle    string   `json:"subtitle"`
	Excerpt     string   `json:"excerpt"`
	Content     string   `json:"content_plain"`
	BlogID      string   `json:"blog_id"`
	BlogHandle  string   `json:"blog_handle"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
	Visibility  string   `json:"visibility"`
	PublishedAt string   `json:"published_at"`
}

// IndexPost upserts one post document.
func (c *Client) IndexPost(ctx context.Context, postID string, doc PostDocument) error {
	status, payload, err := c.do(ctx, http.MethodPut, "/"+c.postsIndex()+"/_doc/"+postID+"?refresh=false", doc)
	if err != nil {
		return err
	}
	if status != http.StatusOK && status != http.StatusCreated {
		return fmt.Errorf("index post %s: status %d: %.200s", postID, status, payload)
	}
	return nil
}

// DeletePost removes one post document. Missing documents are success.
func (c *Client) DeletePost(ctx context.Context, postID string) error {
	status, payload, err := c.do(ctx, http.MethodDelete, "/"+c.postsIndex()+"/_doc/"+postID, nil)
	if err != nil {
		return err
	}
	if status != http.StatusOK && status != http.StatusCreated &&
		status != http.StatusAccepted && status != http.StatusNotFound {
		return fmt.Errorf("delete post %s: status %d: %.200s", postID, status, payload)
	}
	return nil
}

// BlogDocument is the indexed projection of a blog.
type BlogDocument struct {
	Title       string `json:"title"`
	Handle      string `json:"handle"`
	Description string `json:"description"`
}

// IndexBlog upserts one blog document.
func (c *Client) IndexBlog(ctx context.Context, blogID string, doc BlogDocument) error {
	status, payload, err := c.do(ctx, http.MethodPut, "/"+c.blogsIndex()+"/_doc/"+blogID+"?refresh=false", doc)
	if err != nil {
		return err
	}
	if status != http.StatusOK && status != http.StatusCreated {
		return fmt.Errorf("index blog %s: status %d: %.200s", blogID, status, payload)
	}
	return nil
}

// DeleteBlog removes one blog document. Missing documents are success.
func (c *Client) DeleteBlog(ctx context.Context, blogID string) error {
	status, payload, err := c.do(ctx, http.MethodDelete, "/"+c.blogsIndex()+"/_doc/"+blogID, nil)
	if err != nil {
		return err
	}
	if status != http.StatusOK && status != http.StatusCreated &&
		status != http.StatusAccepted && status != http.StatusNotFound {
		return fmt.Errorf("delete blog %s: status %d: %.200s", blogID, status, payload)
	}
	return nil
}

// RankedPost is one ranked search hit.
type RankedPost struct {
	ID    string
	Score float64
}

// SearchPosts runs a ranked multi-match query over published posts.
func (c *Client) SearchPosts(ctx context.Context, query, blogID string, limit, offset int) ([]RankedPost, int, error) {
	must := []map[string]any{
		{"multi_match": map[string]any{
			"query":  query,
			"fields": []string{"title^3", "subtitle^2", "excerpt^2", "content_plain"},
		}},
	}
	filters := []map[string]any{
		{"term": map[string]any{"visibility": "PUBLIC"}},
	}
	if blogID != "" {
		filters = append(filters, map[string]any{"term": map[string]any{"blog_id": blogID}})
	}
	body := map[string]any{
		"from": offset,
		"size": limit,
		"query": map[string]any{
			"bool": map[string]any{"must": must, "filter": filters},
		},
		"_source": false,
		"track_total_hits": true,
	}
	status, payload, err := c.do(ctx, http.MethodPost, "/"+c.postsIndex()+"/_search", body)
	if err != nil {
		return nil, 0, err
	}
	if status != http.StatusOK {
		return nil, 0, fmt.Errorf("search posts: status %d: %.200s", status, payload)
	}
	var parsed struct {
		Hits struct {
			Total struct {
				Value int `json:"value"`
			} `json:"total"`
			Hits []struct {
				ID    string  `json:"_id"`
				Score float64 `json:"_score"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.Unmarshal(payload, &parsed); err != nil {
		return nil, 0, fmt.Errorf("search posts: invalid response: %w", err)
	}
	ranked := make([]RankedPost, 0, len(parsed.Hits.Hits))
	for _, hit := range parsed.Hits.Hits {
		ranked = append(ranked, RankedPost{ID: hit.ID, Score: hit.Score})
	}
	return ranked, parsed.Hits.Total.Value, nil
}
