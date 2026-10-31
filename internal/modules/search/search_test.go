package search

import (
	"context"
	"testing"
)

func TestSimHash(t *testing.T) {
	doc1 := "Go and PostgreSQL high performance blog search engine implementation"
	doc2 := "Go and PostgreSQL high performance blog search engine implementations"
	doc3 := "Completely unrelated text about cooking delicious spaghetti and meatballs"

	h1 := ComputeSimHash64(doc1)
	h2 := ComputeSimHash64(doc2)
	h3 := ComputeSimHash64(doc3)

	dist12 := HammingDistance(h1, h2)
	dist13 := HammingDistance(h1, h3)

	if dist12 > 10 {
		t.Errorf("expected small distance between similar documents, got: %d", dist12)
	}
	if dist13 <= dist12 {
		t.Errorf("expected larger distance between unrelated documents, dist12=%d, dist13=%d", dist12, dist13)
	}
}

func TestSearchFallback(t *testing.T) {
	svc := &Service{}
	res, total, err := svc.SearchPosts(context.Background(), "test", "", 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total == 0 || len(res) == 0 {
		t.Fatalf("expected mock results when pool is nil")
	}
}
