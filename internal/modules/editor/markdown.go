package editor

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// maxConvertInputBytes caps conversion payloads so a single request cannot
// exhaust the converter.
const maxConvertInputBytes = 2_000_000

type mdNode struct {
	Type    string         `json:"type"`
	Text    string         `json:"text"`
	Attrs   map[string]any `json:"attrs"`
	Marks   []mdMark       `json:"marks"`
	Content []mdNode       `json:"content"`
}

type mdMark struct {
	Type  string         `json:"type"`
	Attrs map[string]any `json:"attrs"`
}

// TiptapToMarkdown renders a Tiptap/ProseMirror JSON document to Markdown.
// Unknown node types fall back to rendering their children so content is
// never silently dropped.
func TiptapToMarkdown(astJSON string) (string, error) {
	if len(astJSON) == 0 {
		return "", errors.New("ast payload is required")
	}
	if len(astJSON) > maxConvertInputBytes {
		return "", errors.New("ast payload exceeds size limit")
	}
	var doc mdNode
	dec := json.NewDecoder(strings.NewReader(astJSON))
	if err := dec.Decode(&doc); err != nil {
		return "", fmt.Errorf("invalid ast payload: %w", err)
	}
	var sb strings.Builder
	renderChildren(&sb, doc.Content, "")
	markdown := strings.TrimRight(sb.String(), "\n") + "\n"
	return markdown, nil
}

func renderChildren(sb *strings.Builder, nodes []mdNode, listPrefix string) {
	for i, node := range nodes {
		renderNode(sb, node, listPrefix, i)
	}
}

func renderNode(sb *strings.Builder, node mdNode, listPrefix string, index int) {
	switch node.Type {
	case "paragraph":
		sb.WriteString(listPrefix)
		renderInline(sb, node.Content)
		sb.WriteString("\n")
		if listPrefix == "" {
			sb.WriteString("\n")
		}
	case "heading":
		level := intAttr(node.Attrs, "level", 1)
		if level < 1 {
			level = 1
		}
		if level > 6 {
			level = 6
		}
		sb.WriteString(strings.Repeat("#", level) + " ")
		renderInline(sb, node.Content)
		sb.WriteString("\n\n")
	case "bulletList":
		for _, item := range node.Content {
			renderListItem(sb, item, "- ")
		}
		sb.WriteString("\n")
	case "orderedList":
		start := intAttr(node.Attrs, "start", 1)
		for i, item := range node.Content {
			renderListItem(sb, item, fmt.Sprintf("%d. ", start+i))
		}
		sb.WriteString("\n")
	case "blockquote":
		var inner strings.Builder
		renderChildren(&inner, node.Content, "")
		for _, line := range strings.Split(strings.TrimRight(inner.String(), "\n"), "\n") {
			sb.WriteString("> " + line + "\n")
		}
		sb.WriteString("\n")
	case "codeBlock":
		language, _ := node.Attrs["language"].(string)
		sb.WriteString("```" + language + "\n")
		sb.WriteString(inlineText(node.Content))
		sb.WriteString("\n```\n\n")
	case "horizontalRule":
		sb.WriteString("---\n\n")
	case "image":
		src, _ := node.Attrs["src"].(string)
		alt, _ := node.Attrs["alt"].(string)
		title, _ := node.Attrs["title"].(string)
		if title != "" {
			fmt.Fprintf(sb, "![%s](%s %q)\n\n", alt, src, title)
		} else {
			fmt.Fprintf(sb, "![%s](%s)\n\n", alt, src)
		}
	case "table":
		renderTable(sb, node)
	default:
		// Unknown containers (or a bare doc): render children inline.
		if listPrefix != "" && node.Type == "listItem" {
			renderListItem(sb, node, listPrefix)
			return
		}
		renderChildren(sb, node.Content, listPrefix)
		_ = index
	}
}

func renderListItem(sb *strings.Builder, item mdNode, marker string) {
	if item.Type != "listItem" {
		renderNode(sb, item, "", 0)
		return
	}
	for _, child := range item.Content {
		if child.Type == "paragraph" {
			sb.WriteString(marker)
			renderInline(sb, child.Content)
			sb.WriteString("\n")
		} else if child.Type == "bulletList" || child.Type == "orderedList" {
			var inner strings.Builder
			renderNode(&inner, child, "", 0)
			for _, line := range strings.Split(strings.TrimRight(inner.String(), "\n"), "\n") {
				if strings.TrimSpace(line) == "" {
					continue
				}
				sb.WriteString("  " + line + "\n")
			}
		} else {
			renderNode(sb, child, "", 0)
		}
	}
}

func renderTable(sb *strings.Builder, table mdNode) {
	rows := make([][]string, 0, len(table.Content))
	for _, row := range table.Content {
		if row.Type != "tableRow" {
			continue
		}
		cells := make([]string, 0, len(row.Content))
		for _, cell := range row.Content {
			if cell.Type != "tableCell" && cell.Type != "tableHeader" {
				continue
			}
			var cellSB strings.Builder
			renderInline(&cellSB, cell.Content)
			cells = append(cells, strings.TrimSpace(cellSB.String()))
		}
		if len(cells) > 0 {
			rows = append(rows, cells)
		}
	}
	if len(rows) == 0 {
		return
	}
	width := len(rows[0])
	writeRow := func(cells []string) {
		sb.WriteString("|")
		for i := 0; i < width; i++ {
			cell := ""
			if i < len(cells) {
				cell = cells[i]
			}
			sb.WriteString(" " + cell + " |")
		}
		sb.WriteString("\n")
	}
	writeRow(rows[0])
	sb.WriteString("|")
	for i := 0; i < width; i++ {
		sb.WriteString(" --- |")
	}
	sb.WriteString("\n")
	for _, row := range rows[1:] {
		writeRow(row)
	}
	sb.WriteString("\n")
}

func renderInline(sb *strings.Builder, nodes []mdNode) {
	for _, node := range nodes {
		switch node.Type {
		case "text":
			sb.WriteString(applyMarks(node.Text, node.Marks))
		case "hardBreak":
			sb.WriteString("  \n")
		case "image":
			src, _ := node.Attrs["src"].(string)
			alt, _ := node.Attrs["alt"].(string)
			fmt.Fprintf(sb, "![%s](%s)", alt, src)
		default:
			renderInline(sb, node.Content)
		}
	}
}

func inlineText(nodes []mdNode) string {
	var sb strings.Builder
	for _, node := range nodes {
		if node.Type == "text" {
			sb.WriteString(node.Text)
		} else if node.Type == "hardBreak" {
			sb.WriteString("\n")
		} else {
			sb.WriteString(inlineText(node.Content))
		}
	}
	return sb.String()
}

func applyMarks(text string, marks []mdMark) string {
	var linkHref string
	var bold, italic, strike, code bool
	for _, mark := range marks {
		switch mark.Type {
		case "bold":
			bold = true
		case "italic":
			italic = true
		case "strike":
			strike = true
		case "code":
			code = true
		case "link":
			if href, ok := mark.Attrs["href"].(string); ok && href != "" {
				linkHref = href
			}
		}
	}
	out := text
	if code {
		out = "`" + strings.ReplaceAll(out, "`", "\\`") + "`"
	} else {
		if bold {
			out = "**" + out + "**"
		}
		if italic {
			out = "*" + out + "*"
		}
		if strike {
			out = "~~" + out + "~~"
		}
		if linkHref != "" {
			out = "[" + out + "](" + linkHref + ")"
		}
	}
	return out
}

func intAttr(attrs map[string]any, key string, fallback int) int {
	raw, ok := attrs[key]
	if !ok {
		return fallback
	}
	switch value := raw.(type) {
	case float64:
		return int(value)
	case int:
		return value
	case json.Number:
		if n, err := value.Int64(); err == nil {
			return int(n)
		}
	}
	return fallback
}

// MarkdownToTiptap parses Markdown into a Tiptap/ProseMirror JSON document.
// It supports headings, paragraphs, bullet and ordered lists, blockquotes,
// and fenced code blocks.
func MarkdownToTiptap(markdown string) map[string]any {
	if len(markdown) > maxConvertInputBytes {
		markdown = markdown[:maxConvertInputBytes]
	}
	lines := strings.Split(markdown, "\n")
	content := make([]map[string]any, 0)
	var listItems []map[string]any
	var listOrdered bool
	var listActive bool
	var quoteLines []string
	var codeLines []string
	var codeLanguage string
	inCode := false

	flushList := func() {
		if !listActive || len(listItems) == 0 {
			listActive = false
			listItems = nil
			return
		}
		listType := "bulletList"
		if listOrdered {
			listType = "orderedList"
		}
		content = append(content, map[string]any{"type": listType, "content": listItems})
		listActive = false
		listItems = nil
	}
	flushQuote := func() {
		if len(quoteLines) == 0 {
			return
		}
		blocks := make([]map[string]any, 0, len(quoteLines))
		for _, line := range quoteLines {
			blocks = append(blocks, paragraphNode(line))
		}
		content = append(content, map[string]any{"type": "blockquote", "content": blocks})
		quoteLines = nil
	}
	flushCode := func() {
		if !inCode {
			return
		}
		content = append(content, map[string]any{
			"type":  "codeBlock",
			"attrs": map[string]any{"language": codeLanguage},
			"content": []map[string]any{
				{"type": "text", "text": strings.Join(codeLines, "\n")},
			},
		})
		inCode = false
		codeLines = nil
		codeLanguage = ""
	}

	for _, line := range lines {
		if strings.HasPrefix(line, "```") {
			if inCode {
				flushCode()
			} else {
				flushList()
				flushQuote()
				inCode = true
				codeLanguage = strings.TrimSpace(strings.TrimPrefix(line, "```"))
			}
			continue
		}
		if inCode {
			codeLines = append(codeLines, line)
			continue
		}
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			flushList()
			flushQuote()
			continue
		}
		if strings.HasPrefix(trimmed, ">") {
			flushList()
			quoteLines = append(quoteLines, strings.TrimSpace(strings.TrimPrefix(trimmed, ">")))
			continue
		}
		flushQuote()
		if level, text, ok := parseHeading(trimmed); ok {
			flushList()
			content = append(content, map[string]any{
				"type":  "heading",
				"attrs": map[string]any{"level": level},
				"content": []map[string]any{
					{"type": "text", "text": text},
				},
			})
			continue
		}
		if text, ordered, ok := parseListItem(trimmed); ok {
			if !listActive || listOrdered != ordered {
				flushList()
				listActive = true
				listOrdered = ordered
			}
			listItems = append(listItems, map[string]any{
				"type": "listItem",
				"content": []map[string]any{
					paragraphNode(text),
				},
			})
			continue
		}
		flushList()
		content = append(content, paragraphNode(trimmed))
	}
	flushCode()
	flushList()
	flushQuote()

	return map[string]any{"type": "doc", "content": content}
}

func paragraphNode(text string) map[string]any {
	return map[string]any{
		"type": "paragraph",
		"content": []map[string]any{
			{"type": "text", "text": text},
		},
	}
}

func parseHeading(line string) (int, string, bool) {
	level := 0
	for level < len(line) && line[level] == '#' {
		level++
	}
	if level == 0 || level > 6 || level >= len(line) || line[level] != ' ' {
		return 0, "", false
	}
	return level, strings.TrimSpace(line[level+1:]), true
}

func parseListItem(line string) (string, bool, bool) {
	if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
		return strings.TrimSpace(line[2:]), false, true
	}
	dot := strings.Index(line, ". ")
	if dot > 0 {
		number := strings.TrimSpace(line[:dot])
		ordered := true
		for i := 0; i < len(number); i++ {
			if number[i] < '0' || number[i] > '9' {
				ordered = false
				break
			}
		}
		if ordered {
			return strings.TrimSpace(line[dot+2:]), true, true
		}
	}
	return "", false, false
}
