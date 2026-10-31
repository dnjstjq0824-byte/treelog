package editor

import (
	"context"
	"testing"
)

func TestMarkdownConvert(t *testing.T) {
	svc := &Service{}

	input := MarkdownConvertInput{
		Direction: "TO_AST",
		Markdown:  "# Title\n\n## Subtitle\n\nThis is a paragraph.",
	}

	res, err := svc.ConvertMarkdown(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ast, ok := res["ast"].(map[string]any)
	if !ok {
		t.Fatalf("expected ast in result")
	}

	content, ok := ast["content"].([]map[string]any)
	if !ok || len(content) < 3 {
		t.Fatalf("expected at least 3 content nodes, got: %v", content)
	}
}

func TestLinkPreviewInvalidURL(t *testing.T) {
	svc := &Service{}

	_, err := svc.FetchLinkPreview(context.Background(), "invalid-url")
	if err == nil {
		t.Fatalf("expected error for invalid url")
	}
}
