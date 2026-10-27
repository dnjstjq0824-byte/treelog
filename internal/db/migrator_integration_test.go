package db

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestMigrationsUpDownUp(t *testing.T) {
	url := os.Getenv("TREEBLOG_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TREEBLOG_TEST_DATABASE_URL is not configured")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	pools, err := Open(ctx, url, "")
	if err != nil {
		t.Fatal(err)
	}
	defer pools.Close()
	var databaseName string
	if err := pools.Primary.QueryRow(ctx, "SELECT current_database()").Scan(&databaseName); err != nil {
		t.Fatal(err)
	}
	if databaseName != "treeblog_test" {
		t.Fatalf("refusing migration integration test in database %q", databaseName)
	}
	migrator := Migrator{Pool: pools.Primary}
	files := os.DirFS("../../migrations")
	if err := migrator.Up(ctx, files); err != nil {
		t.Fatal(err)
	}
	if err := migrator.DownOne(ctx, files); err != nil {
		t.Fatal(err)
	}
	if err := migrator.Up(ctx, files); err != nil {
		t.Fatal(err)
	}
	for _, table := range []string{"users", "blogs", "blog_categories"} {
		var exists bool
		if err := pools.Primary.QueryRow(ctx, "SELECT to_regclass($1) IS NOT NULL", table).Scan(&exists); err != nil || !exists {
			t.Fatalf("table %s unavailable: %v", table, err)
		}
	}
}
