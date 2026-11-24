package security

import (
	"strings"
	"testing"
)

func TestSanitizePostHTML(t *testing.T) {
	cases := []struct {
		name         string
		input        string
		mustContain  []string
		mustNotHas   []string
	}{
		{
			name:        "empty input stays empty",
			input:       "",
			mustContain: nil,
			mustNotHas:  nil,
		},
		{
			name:        "script element removed",
			input:       `<p>Hello</p><script>alert(document.cookie)</script>`,
			mustContain: []string{"<p>Hello</p>"},
			mustNotHas:  []string{"<script", "alert("},
		},
		{
			name:        "event handler attributes removed",
			input:       `<p onclick="steal()">Hi</p><img src="https://cdn.example.com/a.png" onerror="steal()" alt="a">`,
			mustContain: []string{"<p>Hi</p>", `src="https://cdn.example.com/a.png"`},
			mustNotHas:  []string{"onclick", "onerror", "steal()"},
		},
		{
			name:        "javascript scheme link neutralized",
			input:       `<a href="javascript:alert(1)">click</a>`,
			mustContain: []string{"click"},
			mustNotHas:  []string{"javascript:"},
		},
		{
			name:        "inline style and style element removed",
			input:       `<style>p{color:expression(alert(1))}</style><p style="color:red">text</p>`,
			mustContain: []string{"text"},
			mustNotHas:  []string{"<style", "style=", "expression("},
		},
		{
			// Bluemonday strips the disallowed src and leaves an inert
			// src-less shell (renders about:blank). The assertion pins
			// the security property: no attacker URL can survive.
			name:        "non allowlisted iframe src stripped",
			input:       `<iframe src="https://evil.example.com/pwn" width="560" height="315"></iframe><p>after</p>`,
			mustContain: []string{"after"},
			mustNotHas:  []string{"evil.example.com", "src="},
		},
		{
			name:        "youtube embed preserved",
			input:       `<iframe src="https://www.youtube.com/embed/dQw4w9WgXcQ" width="560" height="315" title="video" allowfullscreen></iframe>`,
			mustContain: []string{"youtube.com/embed/dQw4w9WgXcQ", "<iframe"},
			mustNotHas:  nil,
		},
		{
			name:        "vimeo embed preserved",
			input:       `<iframe src="https://player.vimeo.com/video/123456" width="560" height="315"></iframe>`,
			mustContain: []string{"player.vimeo.com/video/123456", "<iframe"},
			mustNotHas:  nil,
		},
		{
			name:        "code language class preserved",
			input:       `<pre><code class="language-go">fmt.Println("hi")</code></pre>`,
			mustContain: []string{`class="language-go"`, "fmt.Println"},
			mustNotHas:  nil,
		},
		{
			name:        "arbitrary class stripped",
			input:       `<p class="tb-platform-header">spoof</p>`,
			mustContain: []string{"spoof"},
			mustNotHas:  []string{"tb-platform-header"},
		},
		{
			name:        "rich formatting preserved",
			input:       `<h2>Title</h2><blockquote>quote</blockquote><table><tbody><tr><td>cell</td></tr></tbody></table><ul><li>one</li></ul>`,
			mustContain: []string{"<h2>Title</h2>", "<blockquote>quote</blockquote>", "<table>", "<td>cell</td>", "<li>one</li>"},
			mustNotHas:  nil,
		},
		{
			name:        "form elements removed",
			input:       `<form action="https://evil.example.com"><input type="text" name="x"></form><p>ok</p>`,
			mustContain: []string{"ok"},
			mustNotHas:  []string{"<form", "<input"},
		},
		{
			name:        "svg with script removed",
			input:       `<svg><script>alert(1)</script></svg><p>ok</p>`,
			mustContain: []string{"ok"},
			mustNotHas:  []string{"<svg", "<script"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := SanitizePostHTML(tc.input)
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
