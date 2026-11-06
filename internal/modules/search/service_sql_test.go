package search

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSearchQueriesUseValidProfileColumns guards against SQL runtime errors:
// users has no nickname column (see migrations/0001_identity.up.sql).
// Author names must come from user_profiles.display_name with a fallback
// to users.email when the profile row is missing.
func TestSearchQueriesUseValidProfileColumns(t *testing.T) {
	src, err := os.ReadFile(filepath.Join(".", "service.go"))
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	body := string(src)

	forbidden := []string{"u.nickname", ".nickname"}
	for _, frag := range forbidden {
		if strings.Contains(body, frag) {
			t.Errorf("service.go still references invalid column %q", frag)
		}
	}

	required := []string{"up.display_name", "user_profiles"}
	for _, frag := range required {
		if !strings.Contains(body, frag) {
			t.Errorf("service.go must select author identity from %q", frag)
		}
	}
}
