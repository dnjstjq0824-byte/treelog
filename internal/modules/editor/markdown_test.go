package editor

import (
	"strings"
	"testing"
)

func TestTiptapToMarkdown(t *testing.T) {
	ast := `{"type":"doc","content":[
		{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"제목"}]},
		{"type":"paragraph","content":[
			{"type":"text","text":"굵게","marks":[{"type":"bold"}]},
			{"type":"text","text":"와 "},
			{"type":"text","text":"링크","marks":[{"type":"link","attrs":{"href":"https://example.com"}}]},
			{"type":"text","text":"와 "},
			{"type":"text","text":"코드","marks":[{"type":"code"}]}
		]},
		{"type":"bulletList","content":[
			{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"첫째"}]}]},
			{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"둘째"}]}]}
		]},
		{"type":"blockquote","content":[
			{"type":"paragraph","content":[{"type":"text","text":"인용문"}]}
		]},
		{"type":"codeBlock","attrs":{"language":"go"},"content":[{"type":"text","text":"fmt.Println()"}]},
		{"type":"horizontalRule"}
	]}`
	got, err := TiptapToMarkdown(ast)
	if err != nil {
		t.Fatalf("TiptapToMarkdown: %v", err)
	}
	for _, want := range []string{
		"## 제목",
		"**굵게**와 [링크](https://example.com)와 `코드`",
		"- 첫째",
		"- 둘째",
		"> 인용문",
		"```go",
		"fmt.Println()",
		"---",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}

func TestTiptapToMarkdownTableAndImage(t *testing.T) {
	ast := `{"type":"doc","content":[
		{"type":"image","attrs":{"src":"https://cdn.example.com/a.png","alt":"대체"}},
		{"type":"table","content":[
			{"type":"tableRow","content":[
				{"type":"tableHeader","content":[{"type":"paragraph","content":[{"type":"text","text":"이름"}]}]},
				{"type":"tableHeader","content":[{"type":"paragraph","content":[{"type":"text","text":"나이"}]}]}
			]},
			{"type":"tableRow","content":[
				{"type":"tableCell","content":[{"type":"paragraph","content":[{"type":"text","text":"김트리"}]}]},
				{"type":"tableCell","content":[{"type":"paragraph","content":[{"type":"text","text":"7"}]}]}
			]}
		]}
	]}`
	got, err := TiptapToMarkdown(ast)
	if err != nil {
		t.Fatalf("TiptapToMarkdown: %v", err)
	}
	for _, want := range []string{
		"![대체](https://cdn.example.com/a.png)",
		"| 이름 | 나이 |",
		"| --- | --- |",
		"| 김트리 | 7 |",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}

func TestTiptapToMarkdownRejectsInvalid(t *testing.T) {
	if _, err := TiptapToMarkdown(""); err == nil {
		t.Fatalf("empty payload accepted")
	}
	if _, err := TiptapToMarkdown("{not json"); err == nil {
		t.Fatalf("invalid JSON accepted")
	}
	if _, err := TiptapToMarkdown(strings.Repeat("x", maxConvertInputBytes+1)); err == nil {
		t.Fatalf("oversize payload accepted")
	}
}

func TestMarkdownToTiptap(t *testing.T) {
	doc := MarkdownToTiptap("# 제목\n\n본문 단락\n\n- 첫째\n- 둘째\n\n> 인용\n\n```go\ncode()\n```\n")
	content, ok := doc["content"].([]map[string]any)
	if !ok {
		t.Fatalf("unexpected doc shape: %#v", doc)
	}
	if len(content) != 5 {
		t.Fatalf("expected 5 blocks, got %d", len(content))
	}
	if content[0]["type"] != "heading" {
		t.Errorf("block 0 = %v, want heading", content[0]["type"])
	}
	if content[1]["type"] != "paragraph" {
		t.Errorf("block 1 = %v, want paragraph", content[1]["type"])
	}
	if content[2]["type"] != "bulletList" {
		t.Errorf("block 2 = %v, want bulletList", content[2]["type"])
	}
	if content[3]["type"] != "blockquote" {
		t.Errorf("block 3 = %v, want blockquote", content[3]["type"])
	}
	if content[4]["type"] != "codeBlock" {
		t.Errorf("block 4 = %v, want codeBlock", content[4]["type"])
	}
}

func TestMarkdownRoundTrip(t *testing.T) {
	ast := `{"type":"doc","content":[{"type":"heading","attrs":{"level":1},"content":[{"type":"text","text":"안녕"}]}]}`
	markdown, err := TiptapToMarkdown(ast)
	if err != nil {
		t.Fatalf("TiptapToMarkdown: %v", err)
	}
	if !strings.Contains(markdown, "# 안녕") {
		t.Fatalf("unexpected markdown: %q", markdown)
	}
	back := MarkdownToTiptap(markdown)
	blocks := back["content"].([]map[string]any)
	if len(blocks) != 1 || blocks[0]["type"] != "heading" {
		t.Fatalf("round trip lost the heading: %#v", back)
	}
}
