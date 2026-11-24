package theme

import (
	"strings"
	"testing"
)

func TestSanitizeCSSScopingAndFiltering(t *testing.T) {
	cases := []struct {
		name        string
		input       string
		mustContain []string
		mustNotHas  []string
	}{
		{
			name:        "empty input stays empty",
			input:       "  \n ",
			mustContain: nil,
			mustNotHas:  nil,
		},
		{
			name:        "valid rule preserved and scoped",
			input:       `h1 { color: red; margin: 0 auto; }`,
			mustContain: []string{".tb-blog-theme-root h1", "color: red"},
			mustNotHas:  nil,
		},
		{
			name:        "already scoped selector not double prefixed",
			input:       `.tb-blog-theme-root p { line-height: 1.7; }`,
			mustContain: []string{".tb-blog-theme-root p", "line-height: 1.7"},
			mustNotHas:  []string{".tb-blog-theme-root .tb-blog-theme-root"},
		},
		{
			name:        "root selector rewritten to scope root",
			input:       `:root { --brand: #123456; }`,
			mustContain: []string{".tb-blog-theme-root", "--brand: #123456"},
			mustNotHas:  []string{":root"},
		},
		{
			name:        "platform selectors dropped entirely",
			input:       `.tb-ad-badge { display: none !important; } h2 { color: blue; }`,
			mustContain: []string{".tb-blog-theme-root h2"},
			mustNotHas:  []string{"tb-ad-badge"},
		},
		{
			name:        "mixed selector list keeps only safe selectors",
			input:       `h3, .tb-platform-header { color: green; }`,
			mustContain: []string{".tb-blog-theme-root h3"},
			mustNotHas:  []string{"tb-platform-header"},
		},
		{
			name:        "media query kept with scoped nested rules",
			input:       `@media (min-width: 768px) { .card { display: grid; } }`,
			mustContain: []string{"@media (min-width: 768px)", ".tb-blog-theme-root .card"},
			mustNotHas:  nil,
		},
		{
			name:        "import and charset dropped",
			input:       `@charset "utf-8"; @import url('https://evil.example.com/leak.css'); p { color: black; }`,
			mustContain: []string{".tb-blog-theme-root p"},
			mustNotHas:  []string{"@import", "@charset", "evil.example.com"},
		},
		{
			name:        "keyframes kept with unscoped steps",
			input:       `@keyframes fade { from { opacity: 0; } to { opacity: 1; } }`,
			mustContain: []string{"@keyframes fade", "from", "opacity: 0"},
			mustNotHas:  []string{".tb-blog-theme-root from"},
		},
		{
			name:        "javascript url declaration dropped, siblings kept",
			input:       `.banner { color: white; background-image: url("javascript:alert('XSS')"); }`,
			mustContain: []string{"color: white"},
			mustNotHas:  []string{"javascript:"},
		},
		{
			name:        "data html url dropped, data image kept",
			input:       `.a { background: url(data:text/html,<script>alert(1)</script>); } .b { background: url(data:image/png;base64,iVBOR); }`,
			mustContain: []string{"data:image/png"},
			mustNotHas:  []string{"data:text/html"},
		},
		{
			name:        "relative and https urls kept",
			input:       `.c { background: url(/assets/bg.png); } .d { background: url("https://cdn.example.com/x.jpg"); }`,
			mustContain: []string{"/assets/bg.png", "https://cdn.example.com/x.jpg"},
			mustNotHas:  nil,
		},
		{
			name:        "legacy code execution vectors dropped",
			input:       `body { width: expression(alert(1)); behavior: url(malicious.htc); color: red; }`,
			mustContain: []string{"color: red"},
			mustNotHas:  []string{"expression(", "behavior:"},
		},
		{
			name:        "modern css preserved",
			input:       `.grid { display: grid; grid-template-columns: repeat(2, 1fr); color: var(--brand); }`,
			mustContain: []string{"repeat(2, 1fr)", "var(--brand)"},
			mustNotHas:  nil,
		},
		{
			name:        "style breakout dropped",
			input:       `p::after { content: "</style><script>alert(1)</script>"; color: red; }`,
			mustContain: []string{"color: red"},
			mustNotHas:  []string{"</style", "<script"},
		},
		{
			name:        "rule left with no declarations is dropped",
			input:       `.evil { background: url(javascript:alert(1)); }`,
			mustContain: nil,
			mustNotHas:  []string{".evil", "javascript:"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := SanitizeCSS(tc.input)
			for _, want := range tc.mustContain {
				if !strings.Contains(got, want) {
					t.Errorf("expected output to contain %q, got %q", want, got)
				}
			}
			for _, banned := range tc.mustNotHas {
				if strings.Contains(got, banned) {
					t.Errorf("expected output to exclude %q, got %q", banned, got)
				}
			}
		})
	}
}

func TestSanitizeCSSFailClosed(t *testing.T) {
	// Genuinely unparseable input is rejected wholesale instead of guessed at.
	if got := SanitizeCSS(`<<<< not css >>>>`); got != "" {
		t.Errorf("expected empty output for unparseable CSS, got %q", got)
	}
}

func TestSanitizeCSSLenientRecovery(t *testing.T) {
	// Like browsers, the parser recovers from unclosed blocks; recovered
	// rules still pass through every filter and scoping step.
	got := SanitizeCSS(`h1 { color: red; `)
	if !strings.Contains(got, ".tb-blog-theme-root h1") || !strings.Contains(got, "color: red") {
		t.Errorf("expected recovered scoped rule, got %q", got)
	}
	got = SanitizeCSS(`.tb-ad-badge { display: none; `)
	if strings.Contains(got, "tb-ad-badge") {
		t.Errorf("expected platform selector dropped even in recovered CSS, got %q", got)
	}
}
