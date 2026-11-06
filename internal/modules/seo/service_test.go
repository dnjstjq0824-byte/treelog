package seo

import (
	"context"
	"strings"
	"testing"
)

func TestNewServiceRequiresPool(t *testing.T) {
	if _, err := NewService(nil, "http://x"); err == nil {
		t.Fatalf("nil pool accepted")
	}
}

func TestRobots(t *testing.T) {
	svc := &Service{appBase: "https://blog.example.com"}
	robots := svc.Robots()
	if !strings.Contains(robots, "Sitemap: https://blog.example.com/sitemap.xml") {
		t.Fatalf("robots:\n%s", robots)
	}
}

func TestSitemapUnavailableWithoutPool(t *testing.T) {
	svc := &Service{}
	if _, err := svc.Sitemap(context.Background()); err == nil {
		t.Fatalf("sitemap accepted without pool")
	}
	if _, err := svc.BlogMeta(context.Background(), "h"); err == nil {
		t.Fatalf("meta accepted without pool")
	}
}
