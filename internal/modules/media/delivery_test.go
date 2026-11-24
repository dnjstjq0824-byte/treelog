package media

import (
	"testing"
)

func TestHotlinkAllowed(t *testing.T) {
	svc := &Service{hotlinkHosts: map[string]bool{"cdn.treeblog.example": true}}
	cases := []struct {
		name    string
		referer string
		host    string
		want    bool
	}{
		{"empty referer passes", "", "api.example.com", true},
		{"same host passes", "https://api.example.com/b/x", "api.example.com", true},
		{"same host with ports passes", "http://api.example.com:8080/b/x", "api.example.com:8080", true},
		{"allowlisted host passes", "https://cdn.treeblog.example/img", "api.example.com", true},
		{"foreign host blocked", "https://evil.example.com/", "api.example.com", false},
		{"subdomain spoof blocked", "https://api.example.com.evil.example.com/", "api.example.com", false},
		{"case insensitive", "HTTPS://API.EXAMPLE.COM/", "api.example.com", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := svc.HotlinkAllowed(tc.referer, tc.host); got != tc.want {
				t.Errorf("HotlinkAllowed(%q, %q) = %v, want %v", tc.referer, tc.host, got, tc.want)
			}
		})
	}
}

func TestVariantFormat(t *testing.T) {
	ext, ctype, ok := variantFormat("")
	if !ok || ext != "jpg" || ctype != "image/jpeg" {
		t.Fatalf("default format: %q %q %v", ext, ctype, ok)
	}
	ext, ctype, ok = variantFormat("webp")
	if !ok || ext != "webp" || ctype != "image/webp" {
		t.Fatalf("webp format: %q %q %v", ext, ctype, ok)
	}
	if _, _, ok := variantFormat("avif"); ok {
		t.Fatal("avif must stay rejected until an encoder exists")
	}
}
