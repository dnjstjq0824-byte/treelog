package seo

import (
	"encoding/xml"
	"strings"
	"testing"
	"time"
)

func testFeedChannel() FeedChannel {
	return FeedChannel{
		Title:       "Forest Notes",
		Link:        "https://example.com/b/forest",
		Description: "A test blog",
		Language:    "ko",
		SelfLink:    "https://example.com/b/forest/feed",
		Updated:     time.Date(2026, 11, 24, 9, 30, 0, 0, time.UTC),
	}
}

func testFeedItems() []FeedItem {
	published := time.Date(2026, 11, 23, 12, 0, 0, 0, time.UTC)
	return []FeedItem{
		{
			Title:      `First post <script>alert(1)</script>`,
			Link:       "https://example.com/b/forest/first-post",
			GUID:       "https://example.com/b/forest/first-post",
			Author:     "Alice",
			Published:  published,
			Updated:    published,
			Categories: []string{"essay", "daily"},
			Desc:       "Excerpt with <b>markup</b> & entities.",
			HTML:       "<p>Body</p>",
		},
	}
}

func TestRenderRSS(t *testing.T) {
	body, err := renderRSS(testFeedChannel(), testFeedItems())
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`<rss version="2.0"`,
		"xmlns:content=",
		"<title>Forest Notes</title>",
		"rel=\"self\"",
		"<pubDate>Mon, 23 Nov 2026 12:00:00 +0000</pubDate>",
		"<category>essay</category>",
		"<content:encoded>&lt;p&gt;Body&lt;/p&gt;</content:encoded>",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("expected RSS to contain %q, got:\n%s", want, body)
		}
	}
	// Raw markup from titles and descriptions must be escaped, never emitted.
	for _, banned := range []string{"<script>", "<b>markup</b>", "& entities."} {
		if strings.Contains(body, banned) {
			t.Errorf("expected RSS to escape %q, got:\n%s", banned, body)
		}
	}
	var parsed rssRoot
	if err := xml.Unmarshal([]byte(body), &parsed); err != nil {
		t.Fatalf("RSS output is not well-formed XML: %v\n%s", err, body)
	}
	if len(parsed.Channel.Items) != 1 || parsed.Channel.Items[0].Title != testFeedItems()[0].Title {
		t.Errorf("RSS round-trip lost item data: %+v", parsed.Channel.Items)
	}
}

func TestRenderAtom(t *testing.T) {
	body, err := renderAtom(testFeedChannel(), testFeedItems())
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`<feed xmlns="http://www.w3.org/2005/Atom"`,
		"<title>Forest Notes</title>",
		`rel="self"`,
		"<published>2026-11-23T12:00:00Z</published>",
		`<category term="essay"></category>`,
		`<content type="html">&lt;p&gt;Body&lt;/p&gt;</content>`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("expected Atom to contain %q, got:\n%s", want, body)
		}
	}
	for _, banned := range []string{"<script>", "<b>markup</b>"} {
		if strings.Contains(body, banned) {
			t.Errorf("expected Atom to escape %q, got:\n%s", banned, body)
		}
	}
	var parsed atomRoot
	if err := xml.Unmarshal([]byte(body), &parsed); err != nil {
		t.Fatalf("Atom output is not well-formed XML: %v\n%s", err, body)
	}
	if len(parsed.Entries) != 1 || parsed.Entries[0].ID != testFeedItems()[0].GUID {
		t.Errorf("Atom round-trip lost entry data: %+v", parsed.Entries)
	}
}

func TestRenderFeedsEmpty(t *testing.T) {
	channel := testFeedChannel()
	for _, format := range []string{"rss", "atom"} {
		body, err := renderFeed(format, channel, nil, "")
		if err != nil {
			t.Fatalf("%s: %v", format, err)
		}
		var v any
		switch format {
		case "rss":
			v = &rssRoot{}
		default:
			v = &atomRoot{}
		}
		if err := xml.Unmarshal([]byte(body), v); err != nil {
			t.Errorf("%s: empty feed is not well-formed XML: %v\n%s", format, err, body)
		}
	}
}

func TestRenderFeedSelfLinks(t *testing.T) {
	channel := testFeedChannel()
	body, err := renderFeed("rss", channel, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(body, "https://example.com/b/forest/feed") {
		t.Errorf("expected blog RSS self link, got:\n%s", body)
	}
	body, err = renderFeed("atom", channel, nil, "essay")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(body, "https://example.com/b/forest/tags/essay/atom") {
		t.Errorf("expected tag Atom self link, got:\n%s", body)
	}
}

func TestFeedContentType(t *testing.T) {
	if got := feedContentType("rss"); got != "application/rss+xml; charset=utf-8" {
		t.Errorf("unexpected RSS content type %q", got)
	}
	if got := feedContentType("atom"); got != "application/atom+xml; charset=utf-8" {
		t.Errorf("unexpected Atom content type %q", got)
	}
}
