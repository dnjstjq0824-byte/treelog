package config

import "testing"

func TestLoad(t *testing.T) {
	t.Setenv("TREEBLOG_HTTP_ADDR", "")
	if _, err := Load(); err == nil {
		t.Fatal("expected missing address error")
	}
	t.Setenv("TREEBLOG_HTTP_ADDR", ":8080")
	t.Setenv("TREEBLOG_DATABASE_URL", "postgres://example:example@localhost:5432/treeblog")
	t.Setenv("TREEBLOG_SHUTDOWN_SECONDS", "0")
	if _, err := Load(); err == nil {
		t.Fatal("expected invalid shutdown timeout error")
	}
	t.Setenv("TREEBLOG_SHUTDOWN_SECONDS", "20")
	if _, err := Load(); err != nil {
		t.Fatal(err)
	}
}
