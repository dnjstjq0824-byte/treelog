package theme

import (
	"context"
	"strings"
	"testing"
)

func TestSanitizeCSS(t *testing.T) {
	unsafeCSS := `
		body {
			background: red;
			behavior: url(malicious.htc);
			width: expression(alert(1));
		}
		@import url('https://evil.com/leak.css');
		.banner {
			background-image: url("javascript:alert('XSS')");
		}
		<script>alert(1)</script>
	`

	cleaned := SanitizeCSS(unsafeCSS)

	if strings.Contains(strings.ToLower(cleaned), "javascript:") {
		t.Errorf("SanitizeCSS failed to strip javascript:")
	}
	if strings.Contains(strings.ToLower(cleaned), "@import") {
		t.Errorf("SanitizeCSS failed to strip @import")
	}
	if strings.Contains(strings.ToLower(cleaned), "expression(") {
		t.Errorf("SanitizeCSS failed to strip expression(")
	}
	if strings.Contains(cleaned, "<script>") {
		t.Errorf("SanitizeCSS failed to strip HTML tags")
	}
}

func TestListThemes(t *testing.T) {
	svc := &Service{}
	themes, err := svc.ListThemes(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(themes) == 0 {
		t.Fatalf("expected at least 1 theme in catalog")
	}
}
