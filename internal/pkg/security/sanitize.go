package security

import (
	"regexp"

	"github.com/microcosm-cc/bluemonday"
)

// videoEmbedURL matches only the official embed endpoints of allowlisted
// video providers. Generic iframes are never permitted in post HTML.
var videoEmbedURL = regexp.MustCompile(`^https://(www\.youtube\.com/embed/|youtube\.com/embed/|player\.vimeo\.com/video/)`)

// codeLanguageClass matches Tiptap/lowlight language classes such as
// "language-go" so fenced code blocks survive sanitization.
var codeLanguageClass = regexp.MustCompile(`(?i)^language-[a-z0-9+#_-]+$`)

// postHTMLPolicy is the allowlist HTML policy for persisted post bodies.
// It starts from bluemonday's UGC policy and extends it for the blog editor:
// code-block language classes, image dimension/loading attributes, tables,
// and video embeds restricted to YouTube/Vimeo. Inline styles, scripts,
// event handlers, form elements, and non-allowlisted iframes are stripped.
var postHTMLPolicy = func() *bluemonday.Policy {
	p := bluemonday.UGCPolicy()
	p.AllowAttrs("class").Matching(codeLanguageClass).OnElements("code", "pre")
	p.AllowAttrs("width", "height", "loading", "decoding").OnElements("img")
	p.AllowAttrs("src").Matching(videoEmbedURL).OnElements("iframe")
	p.AllowAttrs("width", "height", "title", "loading", "frameborder", "allow", "allowfullscreen").OnElements("iframe")
	return p
}()

// SanitizePostHTML strips everything outside the post HTML allowlist.
// It is applied to every content_html value before persistence so stored
// post bodies can never carry executable markup, even when the editor
// client is bypassed. An empty input returns an empty string.
func SanitizePostHTML(raw string) string {
	if raw == "" {
		return ""
	}
	return postHTMLPolicy.Sanitize(raw)
}
