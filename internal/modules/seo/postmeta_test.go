package seo

import (
	"encoding/json"
	"testing"
	"time"
)

// TestBuildPostJSONLD pins the BlogPosting + BreadcrumbList schema: required
// Google fields present, valid JSON, breadcrumb chain complete.
func TestBuildPostJSONLD(t *testing.T) {
	published := time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)
	updated := time.Date(2026, 11, 20, 9, 30, 0, 0, time.UTC)
	image := "https://cdn.example.com/cover.jpg"
	blocks := buildPostJSONLD("https://example.com", "forest", "글 제목", "요약",
		"https://example.com/b/forest/slug", &image, "홍길동", &published, updated,
		[]string{"에세이", "일상"})
	if len(blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(blocks))
	}
	encoded, err := json.Marshal(blocks)
	if err != nil {
		t.Fatal(err)
	}
	var decoded []map[string]any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("blocks are not valid JSON: %v", err)
	}
	posting := decoded[0]
	for _, field := range []string{"@context", "@type", "headline", "url", "author", "datePublished", "dateModified", "image", "keywords"} {
		if _, ok := posting[field]; !ok {
			t.Errorf("BlogPosting missing %q", field)
		}
	}
	if posting["@type"] != "BlogPosting" {
		t.Errorf("@type = %v", posting["@type"])
	}
	breadcrumb := decoded[1]
	items, ok := breadcrumb["itemListElement"].([]any)
	if !ok || len(items) != 3 {
		t.Fatalf("breadcrumb items: %+v", breadcrumb["itemListElement"])
	}
	last, _ := items[2].(map[string]any)
	if last["item"] != "https://example.com/b/forest/slug" {
		t.Errorf("breadcrumb leaf = %v", last["item"])
	}
}

// TestBuildPostJSONLDMinimal covers posts without cover, date, or tags.
func TestBuildPostJSONLDMinimal(t *testing.T) {
	updated := time.Date(2026, 11, 20, 9, 30, 0, 0, time.UTC)
	blocks := buildPostJSONLD("https://example.com", "h", "t", "d",
		"https://example.com/b/h/s", nil, "a", nil, updated, nil)
	encoded, _ := json.Marshal(blocks)
	var decoded []map[string]any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if _, ok := decoded[0]["image"]; ok {
		t.Error("image must be absent without a cover")
	}
	if _, ok := decoded[0]["datePublished"]; ok {
		t.Error("datePublished must be absent without a publish date")
	}
}
