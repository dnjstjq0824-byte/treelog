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
	t.Setenv("TREEBLOG_JWT_PRIVATE_KEY_PATH", "test_key.pem")
	t.Setenv("TREEBLOG_CODE_PEPPER", "MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTIzNDU2Nzg5MDE=")
	if _, err := Load(); err != nil {
		t.Fatal(err)
	}
}
