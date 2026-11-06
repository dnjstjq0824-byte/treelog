package post

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestPublicQueriesUseValidProfileColumns guards against SQL runtime errors:
// users has no nickname/handle columns and user_profiles has no avatar_url
// column (see migrations/0001_identity.up.sql). Author identity must come
// from user_profiles.display_name / user_profiles.username with a fallback
// to users.email when the profile row is missing.
func TestPublicQueriesUseValidProfileColumns(t *testing.T) {
	src, err := os.ReadFile(filepath.Join(".", "service.go"))
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	body := string(src)

	forbidden := []string{"u.nickname", "u.handle", "up.avatar_url", "u.avatar_url", ".nickname"}
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
