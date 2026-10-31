package post

import (
	"context"
	"testing"
)

func TestSlugify(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello World", "hello-world"},
		{"안녕하세요 트리블로그 테스트입니다!", "안녕하세요-트리블로그-테스트입니다"},
		{"Special #@$ Characters", "special-characters"},
		{"   leading and trailing spaces   ", "leading-and-trailing-spaces"},
	}

	for _, tt := range tests {
		got := Slugify(tt.input)
		if got != tt.expected {
			t.Errorf("Slugify(%q) = %q, expected %q", tt.input, got, tt.expected)
		}
	}
}

func TestPublishPostValidation(t *testing.T) {
	svc := &Service{}

	_, err := svc.PublishPost(context.Background(), "blog-id", "post-id", "user-id", PublishPostInput{
		Title: "", // Empty title
	})
	if err == nil {
		t.Fatalf("expected error for empty title")
	}
}
