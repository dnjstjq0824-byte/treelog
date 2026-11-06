package search

import (
	"context"
	"errors"
	"net/http"
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

func TestSearchUnavailableWithoutPool(t *testing.T) {
	svc := &Service{}
	_, _, err := svc.SearchPosts(context.Background(), "test", "", 10, 0)
	var fault *Fault
	if !errors.As(err, &fault) {
		t.Fatalf("expected *Fault, got %T (%v)", err, err)
	}
	if fault.Status != http.StatusServiceUnavailable || fault.Code != "SERVICE_UNAVAILABLE" {
		t.Fatalf("expected 503 SERVICE_UNAVAILABLE, got %d %s", fault.Status, fault.Code)
	}
}

func TestCoverImageURL(t *testing.T) {
	if got := coverImageURL(nil); got != nil {
		t.Fatalf("expected nil for nil media id, got %q", *got)
	}
	id := "asset-id"
	got := coverImageURL(&id)
	if got == nil || *got != "/api/v1/public/media/asset-id" {
		t.Fatalf("unexpected cover url: %+v", got)
	}
}

func TestFillTagsGuards(t *testing.T) {
	svc := &Service{}
	if err := svc.fillTags(context.Background(), nil); err != nil {
		t.Fatalf("fillTags empty: %v", err)
	}
	_, _, err := svc.SearchPosts(context.Background(), "test", "", 10, 0)
	var fault *Fault
	if !errors.As(err, &fault) || fault.Code != "SERVICE_UNAVAILABLE" {
		t.Fatalf("expected SERVICE_UNAVAILABLE, got %v", err)
	}
	if err := svc.fillTags(context.Background(), []SearchPostItem{{ID: "p"}}); err == nil {
		t.Fatalf("expected error for fillTags without pool")
	}
}
