package search

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"

	"treeblog/internal/db"
	"treeblog/internal/platform/opensearch"
)

func setupTermsIntegration(t *testing.T) (*Service, func()) {
	t.Helper()
	url := os.Getenv("TREEBLOG_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TREEBLOG_TEST_DATABASE_URL is not configured")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	pools, err := db.Open(ctx, url, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := (db.Migrator{Pool: pools.Primary}).Up(ctx, os.DirFS("../../../migrations")); err != nil {
		pools.Close()
		t.Fatal(err)
	}
	return NewService(pools.Primary, opensearch.New(opensearch.Config{})), func() { pools.Close() }
}

// TestSpamTermLoop covers the G-14 abuse loop: blocking removes a term from
// suggestions, popular lists, and the review queue round-trips the mark.
func TestSpamTermLoop(t *testing.T) {
	svc, cleanup := setupTermsIntegration(t)
	defer cleanup()
	ctx := context.Background()
	prefix := "spamloop-" + uuid.NewString()[:8]
	term := prefix + "-casino"

	if _, err := svc.pool.Exec(ctx, `INSERT INTO search_queries (normalized_query, search_count)
		VALUES ($1, 50)`, term); err != nil {
		t.Fatal(err)
	}
	suggestions, err := svc.Suggest(ctx, prefix, 5)
	if err != nil || len(suggestions) != 1 {
		t.Fatalf("suggest before block: %+v / %v", suggestions, err)
	}
	if err := svc.SetTermBlocked(ctx, term, true); err != nil {
		t.Fatalf("block: %v", err)
	}
	suggestions, _ = svc.Suggest(ctx, prefix, 5)
	for _, s := range suggestions {
		if s == term {
			t.Fatalf("blocked term still suggested: %+v", suggestions)
		}
	}
	popular, err := svc.PopularQueries(ctx, 100)
	if err != nil {
		t.Fatal(err)
	}
	for _, q := range popular {
		if q == term {
			t.Fatal("blocked term still popular")
		}
	}
	blocked, err := svc.ListBlockedTerms(ctx, 100)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, b := range blocked {
		if b == term {
			found = true
		}
	}
	if !found {
		t.Fatal("blocked term missing from review queue")
	}
	if err := svc.SetTermBlocked(ctx, term, false); err != nil {
		t.Fatalf("unblock: %v", err)
	}
	suggestions, _ = svc.Suggest(ctx, prefix, 5)
	if len(suggestions) != 1 || suggestions[0] != term {
		t.Fatalf("unblocked term not restored: %+v", suggestions)
	}
	if err := svc.SetTermBlocked(ctx, "  ", true); err == nil {
		t.Fatal("empty term accepted")
	}
}
