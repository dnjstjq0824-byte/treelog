package post

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCutPaywallPreviewBounds(t *testing.T) {
	doc := json.RawMessage(`{"type":"doc","content":[
		{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"공개 제목"}]},
		{"type":"paragraph","content":[{"type":"text","text":"첫 문단 "},{"type":"text","text":"강조","marks":[{"type":"bold"}]}]},
		{"type":"paragraph","content":[{"type":"text","text":"유료 본문"}]}
	]}`)
	previewJSON, previewHTML := CutPaywallPreview(doc, 2, 500)
	var preview paywallNode
	if err := json.Unmarshal(previewJSON, &preview); err != nil {
		t.Fatal(err)
	}
	if len(preview.Content) != 2 {
		t.Fatalf("expected 2 blocks, got %d: %s", len(preview.Content), previewJSON)
	}
	for _, want := range []string{"<h2>공개 제목</h2>", "<strong>강조</strong>", "paywall-notice"} {
		if !strings.Contains(previewHTML, want) {
			t.Errorf("expected preview HTML to contain %q, got %s", want, previewHTML)
		}
	}
	if strings.Contains(previewHTML, "유료 본문") || strings.Contains(string(previewJSON), "유료 본문") {
		t.Error("paid body leaked past the block cut")
	}
}

func TestCutPaywallPreviewCharCut(t *testing.T) {
	doc := json.RawMessage(`{"type":"doc","content":[
		{"type":"paragraph","content":[{"type":"text","text":"0123456789ABCDEF"}]}
	]}`)
	previewJSON, previewHTML := CutPaywallPreview(doc, 2, 10)
	if strings.Contains(string(previewJSON), "ABCDEF") || strings.Contains(previewHTML, "ABCDEF") {
		t.Error("text leaked past the character cut")
	}
	if !strings.Contains(previewHTML, "0123456789") {
		t.Errorf("expected kept prefix, got %s", previewHTML)
	}
}

func TestCutPaywallPreviewEscapes(t *testing.T) {
	doc := json.RawMessage(`{"type":"doc","content":[
		{"type":"paragraph","content":[
			{"type":"text","text":"<script>alert(1)</script>"},
			{"type":"text","text":"evil","marks":[{"type":"link","attrs":{"href":"javascript:alert(1)"}}]},
			{"type":"text","text":"ok","marks":[{"type":"link","attrs":{"href":"https://example.com/x"}}]}
		]},
		{"type":"mystery","content":[{"type":"text","text":"dropped"}]}
	]}`)
	previewJSON, previewHTML := CutPaywallPreview(doc, 2, 500)
	for _, banned := range []string{"<script>", "javascript:", "mystery", "dropped"} {
		if strings.Contains(previewHTML, banned) || strings.Contains(string(previewJSON), banned) {
			t.Errorf("expected %q excluded, got html=%s json=%s", banned, previewHTML, previewJSON)
		}
	}
	if !strings.Contains(previewHTML, `href="https://example.com/x"`) {
		t.Errorf("expected safe link kept, got %s", previewHTML)
	}
	if !strings.Contains(previewHTML, "&lt;script&gt;") {
		t.Errorf("expected escaped text, got %s", previewHTML)
	}
}

func TestCutPaywallPreviewDegenerate(t *testing.T) {
	for _, input := range []string{"", "{broken", `{"type":"paragraph"}`, `{"type":"doc","content":[]}`} {
		previewJSON, previewHTML := CutPaywallPreview(json.RawMessage(input), 2, 500)
		var preview paywallNode
		if err := json.Unmarshal(previewJSON, &preview); err != nil || preview.Type != "doc" {
			t.Errorf("input %q: invalid preview doc %s", input, previewJSON)
		}
		if !strings.Contains(previewHTML, "paywall-notice") {
			t.Errorf("input %q: notice missing: %s", input, previewHTML)
		}
	}
	previewJSON, previewHTML := CutPaywallPreview(json.RawMessage(`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"x"}]}]}`), 0, 0)
	if strings.Contains(string(previewJSON), `"text":"x"`) || strings.Contains(previewHTML, ">x<") {
		t.Error("zero budget must cut everything")
	}
}
