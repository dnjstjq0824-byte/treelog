package post

import (
	"encoding/json"
	"html"
	"strings"
)

// Paywall preview bounds: at most two leading blocks and 500 characters of
// text. Everything past the cut stays server-side.
const (
	paywallPreviewBlocks = 2
	paywallPreviewChars  = 500
	paywallNoticeText    = "유료 멤버십 또는 단건 결제 후 열람 가능한 콘텐츠입니다."
)

// paywallNode mirrors the editor AST shape for preview cutting.
type paywallNode struct {
	Type    string            `json:"type"`
	Attrs   map[string]any    `json:"attrs,omitempty"`
	Content []paywallNode     `json:"content,omitempty"`
	Marks   []paywallMark     `json:"marks,omitempty"`
	Text    string            `json:"text,omitempty"`
}

// paywallMark mirrors a text mark.
type paywallMark struct {
	Type  string         `json:"type"`
	Attrs map[string]any `json:"attrs,omitempty"`
}

// CutPaywallPreview slices the leading blocks of a post AST into a bounded
// preview document plus safely rendered HTML. Unknown node types are
// dropped (their text is preserved as plain paragraphs only when they carry
// none of their own markup), and every string is escaped, so preview output
// can never smuggle markup or leak past the cut.
func CutPaywallPreview(docJSON json.RawMessage, maxBlocks, maxChars int) (previewJSON json.RawMessage, previewHTML string) {
	preview := cutAST(docJSON, maxBlocks, maxChars)
	encoded, err := json.Marshal(preview)
	if err != nil {
		encoded = []byte(`{"type":"doc","content":[]}`)
	}
	var html strings.Builder
	html.WriteString("<div class=\"paywall-preview\">")
	for _, block := range preview.Content {
		html.WriteString(renderPreviewNode(block))
	}
	html.WriteString("<div class=\"paywall-notice\">" + htmlEscape(paywallNoticeText) + "</div></div>")
	return encoded, html.String()
}

func cutAST(docJSON json.RawMessage, maxBlocks, maxChars int) paywallNode {
	preview := paywallNode{Type: "doc"}
	if maxBlocks <= 0 || maxChars <= 0 {
		return preview
	}
	var doc paywallNode
	if err := json.Unmarshal(docJSON, &doc); err != nil || doc.Type != "doc" {
		return preview
	}
	remaining := maxChars
	for _, block := range doc.Content {
		if len(preview.Content) >= maxBlocks || remaining <= 0 {
			break
		}
		cut, used := cutNode(block, remaining)
		if cut == nil {
			continue
		}
		remaining -= used
		preview.Content = append(preview.Content, *cut)
	}
	return preview
}

// cutNode bounds one block to the remaining character budget. It returns the
// cut node and the characters consumed, or nil when the block contributes
// nothing renderable.
func cutNode(node paywallNode, budget int) (*paywallNode, int) {
	switch node.Type {
	case "paragraph", "heading", "blockquote", "codeBlock", "listItem", "tableCell", "tableHeader":
		out := paywallNode{Type: node.Type, Attrs: filterPreviewAttrs(node.Type, node.Attrs)}
		used := 0
		for _, child := range node.Content {
			if used >= budget {
				break
			}
			cut, n := cutInline(child, budget-used)
			if cut != nil {
				out.Content = append(out.Content, *cut)
				used += n
			}
		}
		if len(out.Content) == 0 {
			return nil, 0
		}
		return &out, used
	case "bulletList", "orderedList", "table", "tableRow":
		out := paywallNode{Type: node.Type}
		used := 0
		for _, child := range node.Content {
			if used >= budget {
				break
			}
			cut, n := cutNode(child, budget-used)
			if cut != nil {
				out.Content = append(out.Content, *cut)
				used += n
			}
		}
		if len(out.Content) == 0 {
			return nil, 0
		}
		return &out, used
	case "image":
		src, _ := node.Attrs["src"].(string)
		if !allowedPreviewURL(src) {
			return nil, 0
		}
		alt, _ := node.Attrs["alt"].(string)
		return &paywallNode{Type: "image", Attrs: map[string]any{"src": src, "alt": alt}}, 0
	case "horizontalRule", "hardBreak":
		return &paywallNode{Type: node.Type}, 0
	default:
		return nil, 0
	}
}

// cutInline bounds inline content (text and its marks).
func cutInline(node paywallNode, budget int) (*paywallNode, int) {
	if node.Type == "hardBreak" {
		return &paywallNode{Type: "hardBreak"}, 0
	}
	if node.Type != "text" {
		return nil, 0
	}
	runes := []rune(node.Text)
	if len(runes) > budget {
		runes = runes[:budget]
	}
	if len(runes) == 0 {
		return nil, 0
	}
	out := paywallNode{Type: "text", Text: string(runes)}
	for _, mark := range node.Marks {
		switch mark.Type {
		case "bold", "italic", "strike", "code", "underline":
			out.Marks = append(out.Marks, paywallMark{Type: mark.Type})
		case "link":
			href, _ := mark.Attrs["href"].(string)
			if allowedPreviewURL(href) {
				out.Marks = append(out.Marks, paywallMark{Type: "link", Attrs: map[string]any{"href": href}})
			}
		}
	}
	return &out, len(runes)
}

func filterPreviewAttrs(nodeType string, attrs map[string]any) map[string]any {
	if nodeType == "heading" {
		if level, ok := attrs["level"].(float64); ok && level >= 1 && level <= 6 {
			return map[string]any{"level": level}
		}
		return nil
	}
	if nodeType == "codeBlock" {
		if language, ok := attrs["language"].(string); ok && language != "" {
			return map[string]any{"language": language}
		}
	}
	return nil
}

func allowedPreviewURL(raw string) bool {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return false
	}
	lower := strings.ToLower(trimmed)
	return strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") ||
		strings.HasPrefix(trimmed, "/") || strings.HasPrefix(trimmed, "#") ||
		strings.HasPrefix(lower, "mailto:")
}

func htmlEscape(s string) string {
	return html.EscapeString(s)
}

// renderPreviewNode renders one cut block. Only the allowlisted shapes
// produced by cutNode arrive here; anything else renders as nothing.
func renderPreviewNode(node paywallNode) string {
	var b strings.Builder
	switch node.Type {
	case "paragraph":
		b.WriteString("<p>")
		for _, child := range node.Content {
			b.WriteString(renderPreviewInline(child))
		}
		b.WriteString("</p>")
	case "heading":
		level := 2
		if l, ok := node.Attrs["level"].(float64); ok && l >= 1 && l <= 6 {
			level = int(l)
		}
		b.WriteString("<h" + string(rune('0'+level)) + ">")
		for _, child := range node.Content {
			b.WriteString(renderPreviewInline(child))
		}
		b.WriteString("</h" + string(rune('0'+level)) + ">")
	case "blockquote":
		b.WriteString("<blockquote>")
		for _, child := range node.Content {
			b.WriteString(renderPreviewInline(child))
		}
		b.WriteString("</blockquote>")
	case "codeBlock":
		b.WriteString("<pre><code>")
		for _, child := range node.Content {
			if child.Type == "text" {
				b.WriteString(htmlEscape(child.Text))
			}
		}
		b.WriteString("</code></pre>")
	case "bulletList":
		b.WriteString("<ul>")
		for _, child := range node.Content {
			b.WriteString(renderPreviewNode(child))
		}
		b.WriteString("</ul>")
	case "orderedList":
		b.WriteString("<ol>")
		for _, child := range node.Content {
			b.WriteString(renderPreviewNode(child))
		}
		b.WriteString("</ol>")
	case "listItem":
		b.WriteString("<li>")
		for _, child := range node.Content {
			b.WriteString(renderPreviewInline(child))
		}
		b.WriteString("</li>")
	case "table":
		b.WriteString("<table><tbody>")
		for _, child := range node.Content {
			b.WriteString(renderPreviewNode(child))
		}
		b.WriteString("</tbody></table>")
	case "tableRow":
		b.WriteString("<tr>")
		for _, child := range node.Content {
			b.WriteString(renderPreviewNode(child))
		}
		b.WriteString("</tr>")
	case "tableCell", "tableHeader":
		tag := "td"
		if node.Type == "tableHeader" {
			tag = "th"
		}
		b.WriteString("<" + tag + ">")
		for _, child := range node.Content {
			b.WriteString(renderPreviewInline(child))
		}
		b.WriteString("</" + tag + ">")
	case "image":
		src, _ := node.Attrs["src"].(string)
		alt, _ := node.Attrs["alt"].(string)
		b.WriteString("<img src=\"" + htmlEscape(src) + "\" alt=\"" + htmlEscape(alt) + "\" loading=\"lazy\">")
	case "horizontalRule":
		b.WriteString("<hr>")
	}
	return b.String()
}

func renderPreviewInline(node paywallNode) string {
	if node.Type == "hardBreak" {
		return "<br>"
	}
	if node.Type != "text" {
		return ""
	}
	text := htmlEscape(node.Text)
	for i := len(node.Marks) - 1; i >= 0; i-- {
		switch node.Marks[i].Type {
		case "bold":
			text = "<strong>" + text + "</strong>"
		case "italic":
			text = "<em>" + text + "</em>"
		case "strike":
			text = "<s>" + text + "</s>"
		case "code":
			text = "<code>" + text + "</code>"
		case "underline":
			text = "<u>" + text + "</u>"
		case "link":
			href, _ := node.Marks[i].Attrs["href"].(string)
			text = "<a href=\"" + htmlEscape(href) + "\" rel=\"noopener nofollow\">" + text + "</a>"
		}
	}
	return text
}
