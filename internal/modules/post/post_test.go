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

func TestCoverImageURL(t *testing.T) {
	if got := coverImageURL(nil); got != nil {
		t.Fatalf("expected nil for nil media id, got %q", *got)
	}
	empty := ""
	if got := coverImageURL(&empty); got != nil {
		t.Fatalf("expected nil for empty media id, got %q", *got)
	}
	id := "9f3a1c2e-7b4d-4a1f-8e2c-5d6b7a8c9d0e"
	got := coverImageURL(&id)
	if got == nil || *got != "/api/v1/public/media/"+id {
		t.Fatalf("unexpected cover url: %+v", got)
	}
}

func TestManageNilPool(t *testing.T) {
	svc := &Service{}
	ctx := context.Background()
	if _, err := svc.ListManagePosts(ctx, "b", "u", "", 20, 0); !isServiceUnavailable(err) {
		t.Fatalf("ListManagePosts without pool: %v", err)
	}
	if _, err := svc.GetManagePost(ctx, "b", "p", "u"); !isServiceUnavailable(err) {
		t.Fatalf("GetManagePost without pool: %v", err)
	}
	if err := svc.fillSummaryTags(ctx, nil); err != nil {
		t.Fatalf("fillSummaryTags empty: %v", err)
	}
	if err := svc.fillSummaryTags(ctx, []PostSummary{{ID: "p"}}); !isServiceUnavailable(err) {
		t.Fatalf("fillSummaryTags without pool: %v", err)
	}
}

func isServiceUnavailable(err error) bool {
	fault, ok := err.(*Fault)
	return ok && fault != nil && fault.Status == 503 && fault.Code == "SERVICE_UNAVAILABLE"
}
