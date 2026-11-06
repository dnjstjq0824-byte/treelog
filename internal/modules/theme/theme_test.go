package theme

import (
	"context"
	"errors"
	"net/http"
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

func TestListThemesUnavailableWithoutPool(t *testing.T) {
	svc := &Service{}
	_, err := svc.ListThemes(context.Background())
	var fault *Fault
	if !errors.As(err, &fault) {
		t.Fatalf("expected *Fault, got %T (%v)", err, err)
	}
	if fault.Status != http.StatusServiceUnavailable || fault.Code != "SERVICE_UNAVAILABLE" {
		t.Fatalf("expected 503 SERVICE_UNAVAILABLE, got %d %s", fault.Status, fault.Code)
	}
}
