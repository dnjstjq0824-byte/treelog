package seo

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"treeblog/internal/pkg/security"
)

// feedLimit caps feed entries per response.
const feedLimit = 20

// FeedChannel describes the blog a feed belongs to.
type FeedChannel struct {
	Title       string
	Link        string
	Description string
	Language    string
	SelfLink    string
	Updated     time.Time
}

// FeedItem is one feed entry. Description is plain text; HTML carries the
// sanitized post body and is empty for comment entries.
type FeedItem struct {
	Title      string
	Link       string
	GUID       string
	Author     string
	Published  time.Time
	Updated    time.Time
	Categories []string
	Desc       string
	HTML       string
}

type rssRoot struct {
	XMLName   xml.Name   `xml:"rss"`
	Version   string     `xml:"version,attr"`
	ContentNS string     `xml:"xmlns:content,attr"`
	AtomNS    string     `xml:"xmlns:atom,attr"`
	Channel   rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Language    string    `xml:"language,omitempty"`
	Updated     string    `xml:"lastBuildDate"`
	SelfLink    rssAtomLink `xml:"atom:link"`
	Items       []rssItem `xml:"item"`
}

type rssAtomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
}

type rssItem struct {
	Title       string   `xml:"title"`
	Link        string   `xml:"link"`
	GUID        string   `xml:"guid"`
	Author      string   `xml:"author,omitempty"`
	Published   string   `xml:"pubDate"`
	Categories  []string `xml:"category"`
	Description string   `xml:"description"`
	Content     string   `xml:"content:encoded,omitempty"`
}

type atomRoot struct {
	XMLName   xml.Name    `xml:"feed"`
	Xmlns     string      `xml:"xmlns,attr"`
	Title     string      `xml:"title"`
	Subtitle  string      `xml:"subtitle,omitempty"`
	Links     []atomLink  `xml:"link"`
	Updated   string      `xml:"updated"`
	ID        string      `xml:"id"`
	Entries   []atomEntry `xml:"entry"`
}

type atomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr,omitempty"`
	Type string `xml:"type,attr,omitempty"`
}

type atomEntry struct {
	Title     string          `xml:"title"`
	Link      atomLink        `xml:"link"`
	ID        string          `xml:"id"`
	Published string          `xml:"published"`
	Updated   string          `xml:"updated"`
	Author    atomAuthor      `xml:"author"`
	Cats      []atomCategory `xml:"category"`
	Summary   string          `xml:"summary,omitempty"`
	Content   atomContent     `xml:"content"`
}

type atomAuthor struct {
	Name string `xml:"name"`
}

type atomCategory struct {
	Term string `xml:"term,attr"`
}

type atomContent struct {
	Type string `xml:"type,attr"`
	Body string `xml:",chardata"`
}

// renderRSS builds an RSS 2.0 document. encoding/xml escapes every string,
// so titles and descriptions can never break feed markup.
func renderRSS(channel FeedChannel, items []FeedItem) (string, error) {
	updated := channel.Updated
	if updated.IsZero() {
		updated = time.Now().UTC()
	}
	out := rssRoot{
		Version:   "2.0",
		ContentNS: "http://purl.org/rss/1.0/modules/content/",
		AtomNS:    "http://www.w3.org/2005/Atom",
		Channel: rssChannel{
			Title:       channel.Title,
			Link:        channel.Link,
			Description: channel.Description,
			Language:    channel.Language,
			Updated:     updated.Format(time.RFC1123Z),
			SelfLink:    rssAtomLink{Href: channel.SelfLink, Rel: "self", Type: "application/rss+xml"},
		},
	}
	for _, item := range items {
		published := item.Published
		if published.IsZero() {
			published = updated
		}
		out.Channel.Items = append(out.Channel.Items, rssItem{
			Title:       item.Title,
			Link:        item.Link,
			GUID:        item.GUID,
			Author:      item.Author,
			Published:   published.Format(time.RFC1123Z),
			Categories:  item.Categories,
			Description: item.Desc,
			Content:     item.HTML,
		})
	}
	encoded, err := xml.MarshalIndent(out, "", "  ")
	if err != nil {
		return "", err
	}
	return xml.Header + string(encoded), nil
}

// renderAtom builds an RFC 4287 Atom 1.0 document.
func renderAtom(channel FeedChannel, items []FeedItem) (string, error) {
	updated := channel.Updated
	if updated.IsZero() {
		updated = time.Now().UTC()
	}
	out := atomRoot{
		Xmlns:    "http://www.w3.org/2005/Atom",
		Title:    channel.Title,
		Subtitle: channel.Description,
		Links: []atomLink{
			{Href: channel.Link},
			{Href: channel.SelfLink, Rel: "self", Type: "application/atom+xml"},
		},
		Updated: updated.Format(time.RFC3339),
		ID:      channel.Link,
	}
	for _, item := range items {
		published := item.Published
		if published.IsZero() {
			published = updated
		}
		entryUpdated := item.Updated
		if entryUpdated.IsZero() {
			entryUpdated = published
		}
		entry := atomEntry{
			Title:     item.Title,
			Link:      atomLink{Href: item.Link},
			ID:        item.GUID,
			Published: published.Format(time.RFC3339),
			Updated:   entryUpdated.Format(time.RFC3339),
			Author:    atomAuthor{Name: item.Author},
			Summary:   item.Desc,
			Content:   atomContent{Type: "html", Body: item.HTML},
		}
		for _, cat := range item.Categories {
			entry.Cats = append(entry.Cats, atomCategory{Term: cat})
		}
		out.Entries = append(out.Entries, entry)
	}
	encoded, err := xml.MarshalIndent(out, "", "  ")
	if err != nil {
		return "", err
	}
	return xml.Header + string(encoded), nil
}

// feedBlog resolves an active blog handle to its id and display fields.
func (s *Service) feedBlog(ctx context.Context, handle string) (id, title, description string, err error) {
	if err := s.requirePool(); err != nil {
		return "", "", "", err
	}
	err = s.pool.QueryRow(ctx, `
		SELECT id::text, title, COALESCE(description, '')
		FROM blogs WHERE handle = $1 AND status = 'ACTIVE'
	`, handle).Scan(&id, &title, &description)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", "", &Fault{Status: http.StatusNotFound, Code: "BLOG_NOT_FOUND", Message: "Blog not found"}
	}
	if err != nil {
		return "", "", "", err
	}
	if description == "" {
		description = title
	}
	return id, title, description, nil
}

// BlogPostFeed renders the latest public posts of a blog, optionally filtered
// by tag slug, in RSS ("rss") or Atom ("atom") format.
func (s *Service) BlogPostFeed(ctx context.Context, handle, tag, format string) (string, error) {
	blogID, title, description, err := s.feedBlog(ctx, handle)
	if err != nil {
		return "", err
	}
	args := []any{blogID}
	tagFilter := ""
	if tag != "" {
		tagFilter = `AND EXISTS (
			SELECT 1 FROM post_tags pt JOIN tags t ON t.id = pt.tag_id
			WHERE pt.post_id = p.id AND t.slug = $2
		)`
		args = append(args, strings.ToLower(tag))
	}
	rows, err := s.pool.Query(ctx, `
		SELECT p.slug, p.title, COALESCE(p.excerpt, ''), COALESCE(p.content_html, ''),
		       COALESCE(up.display_name, u.email::text),
		       COALESCE(p.published_at, p.created_at), p.updated_at,
		       COALESCE(array_agg(t.name) FILTER (WHERE t.name IS NOT NULL), '{}')
		FROM posts p
		JOIN users u ON u.id = p.author_user_id
		LEFT JOIN user_profiles up ON up.user_id = u.id
		LEFT JOIN post_tags pt ON pt.post_id = p.id
		LEFT JOIN tags t ON t.id = pt.tag_id
		WHERE p.blog_id = $1::uuid AND p.status = 'PUBLISHED' AND p.visibility = 'PUBLIC'
		  AND p.trashed_at IS NULL `+tagFilter+`
		GROUP BY p.id, p.slug, p.title, p.excerpt, p.content_html, up.display_name,
		         u.email, p.published_at, p.created_at, p.updated_at
		ORDER BY COALESCE(p.published_at, p.created_at) DESC
		LIMIT `+fmt.Sprintf("%d", feedLimit),
		args...)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	blogLink := s.appBase + "/b/" + handle
	channel := FeedChannel{Title: title, Link: blogLink, Description: description, Language: "ko"}
	items := make([]FeedItem, 0, feedLimit)
	for rows.Next() {
		var slug, itemTitle, excerpt, htmlBody, author string
		var published, updated time.Time
		var cats []string
		if err := rows.Scan(&slug, &itemTitle, &excerpt, &htmlBody, &author, &published, &updated, &cats); err != nil {
			return "", err
		}
		link := blogLink + "/" + slug
		items = append(items, FeedItem{
			Title:      itemTitle,
			Link:       link,
			GUID:       link,
			Author:     author,
			Published:  published,
			Updated:    updated,
			Categories: cats,
			Desc:       excerpt,
			// Defense in depth: rows written before server-side
			// sanitization are cleaned again at feed render time.
			HTML: security.SanitizePostHTML(htmlBody),
		})
	}
	if err := rows.Err(); err != nil {
		return "", err
	}
	if len(items) > 0 {
		channel.Updated = items[0].Published
	}
	return renderFeed(format, channel, items, tag)
}

// BlogCommentFeed renders recent public, non-secret comments of a blog.
func (s *Service) BlogCommentFeed(ctx context.Context, handle, format string) (string, error) {
	blogID, title, _, err := s.feedBlog(ctx, handle)
	if err != nil {
		return "", err
	}
	rows, err := s.pool.Query(ctx, `
		SELECT c.id::text, p.slug, p.title, COALESCE(up.display_name, u.email::text),
		       c.content_text, c.created_at
		FROM comments c
		JOIN posts p ON p.id = c.post_id
		JOIN users u ON u.id = c.author_user_id
		LEFT JOIN user_profiles up ON up.user_id = u.id
		WHERE c.blog_id = $1::uuid AND c.status = 'PUBLISHED' AND c.is_secret = FALSE
		  AND p.status = 'PUBLISHED' AND p.visibility = 'PUBLIC' AND p.trashed_at IS NULL
		ORDER BY c.created_at DESC
		LIMIT `+fmt.Sprintf("%d", feedLimit),
		blogID)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	blogLink := s.appBase + "/b/" + handle
	channel := FeedChannel{
		Title:       title + " - 최근 댓글",
		Link:        blogLink,
		Description: title + " 블로그의 최근 댓글",
		Language:    "ko",
	}
	items := make([]FeedItem, 0, feedLimit)
	for rows.Next() {
		var commentID, slug, postTitle, author, body string
		var created time.Time
		if err := rows.Scan(&commentID, &slug, &postTitle, &author, &body, &created); err != nil {
			return "", err
		}
		link := blogLink + "/" + slug + "#comment-" + commentID
		items = append(items, FeedItem{
			Title:     "“" + postTitle + "”에 대한 " + author + "님의 댓글",
			Link:      link,
			GUID:      link,
			Author:    author,
			Published: created,
			Updated:   created,
			Desc:      body,
		})
	}
	if err := rows.Err(); err != nil {
		return "", err
	}
	if len(items) > 0 {
		channel.Updated = items[0].Published
	}
	return renderFeed(format, channel, items, "")
}

func renderFeed(format string, channel FeedChannel, items []FeedItem, tag string) (string, error) {
	self := channel.Link + "/feed"
	if format == "atom" {
		self = channel.Link + "/atom"
	}
	if tag != "" {
		self = channel.Link + "/tags/" + tag + "/" + format
		if format == "rss" {
			self = channel.Link + "/tags/" + tag + "/feed"
		}
	}
	channel.SelfLink = self
	if format == "atom" {
		return renderAtom(channel, items)
	}
	return renderRSS(channel, items)
}
