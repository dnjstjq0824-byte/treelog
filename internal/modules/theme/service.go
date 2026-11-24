package theme

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/aymerick/douceur/css"
	"github.com/aymerick/douceur/parser"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Fault struct {
	Status  int
	Code    string
	Message string
}

func (f *Fault) Error() string {
	return f.Message
}

type ThemeCatalogItem struct {
	ID                string          `json:"id"`
	Slug              string          `json:"slug"`
	Name              string          `json:"name"`
	Description       string          `json:"description"`
	LayoutType        string          `json:"layout_type"`
	DefaultTokensJSON json.RawMessage `json:"default_tokens_json"`
	PreviewImageURL   *string         `json:"preview_image_url,omitempty"`
	Version           string          `json:"version"`
	IsActive          bool            `json:"is_active"`
	CreatedAt         time.Time       `json:"created_at"`
}

type BlogThemeConfig struct {
	BlogID             string          `json:"blog_id"`
	ThemeID            string          `json:"theme_id"`
	ThemeSlug          string          `json:"theme_slug"`
	ThemeName          string          `json:"theme_name"`
	ActiveLayout       string          `json:"active_layout"`
	ColorPaletteTokens json.RawMessage `json:"color_palette_tokens"`
	TypographyTokens   json.RawMessage `json:"typography_tokens"`
	CustomCSS          string          `json:"custom_css"`
	DarkModePreference string          `json:"dark_mode_preference"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

type UpdateBlogThemeInput struct {
	ThemeID            *string          `json:"theme_id,omitempty"`
	ActiveLayout       *string          `json:"active_layout,omitempty"`
	ColorPaletteTokens *json.RawMessage `json:"color_palette_tokens,omitempty"`
	TypographyTokens   *json.RawMessage `json:"typography_tokens,omitempty"`
	CustomCSS          *string          `json:"custom_css,omitempty"`
	DarkModePreference *string          `json:"dark_mode_preference,omitempty"`
}

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

// themeScopeRoot confines every custom CSS selector to the blog theme subtree.
const themeScopeRoot = ".tb-blog-theme-root"

// platformSelectorFragments are platform-owned UI hooks that custom CSS must
// never match, so users cannot hide or spoof the platform header, ad badge,
// or report button.
var platformSelectorFragments = []string{".tb-platform-header", ".tb-ad-badge", ".tb-report-btn"}

// allowedAtRules is the closed set of at-rules permitted in custom CSS.
// Everything else (@import, @charset, @namespace, @document, @page, ...) is
// dropped without parsing its payload.
var allowedAtRules = map[string]bool{
	"@media":     true,
	"@supports":  true,
	"@keyframes": true,
	"@font-face": true,
}

// blockedDeclarationProps are legacy properties that execute code in old
// engines and have no legitimate use in blog themes.
var blockedDeclarationProps = map[string]bool{
	"behavior":    true,
	"-moz-binding": true,
}

// blockedValueFragments are case-insensitive substrings that never appear in
// legitimate declaration values.
var blockedValueFragments = []string{
	"javascript:",
	"expression(",
	"behavior:",
	"-moz-binding",
	"</style",
}

var cssURLTarget = regexp.MustCompile(`(?i)url\(\s*['"]?\s*([^'")\s]+)`)

// SanitizeCSS parses custom theme CSS and re-emits only the safe subset:
// disallowed at-rules, platform selectors, dangerous declarations, and
// unsafe url() targets are dropped, and every remaining selector is scoped
// under .tb-blog-theme-root. Genuinely unparseable input yields an empty
// string (fail closed); recoverable slips such as unclosed blocks are
// parsed leniently like browsers do, with every recovered rule filtered.
func SanitizeCSS(rawCSS string) string {
	if strings.TrimSpace(rawCSS) == "" {
		return ""
	}
	sheet, err := parser.Parse(rawCSS)
	if err != nil {
		return ""
	}
	out := css.NewStylesheet()
	for _, rule := range sheet.Rules {
		if clean := sanitizeCSSRule(rule, true); clean != nil {
			out.Rules = append(out.Rules, clean)
		}
	}
	return out.String()
}

func sanitizeCSSRule(rule *css.Rule, scope bool) *css.Rule {
	if rule.Kind == css.AtRule {
		return sanitizeCSSAtRule(rule)
	}
	selectors := make([]string, 0, len(rule.Selectors))
	for _, sel := range rule.Selectors {
		sel = strings.TrimSpace(sel)
		if sel == "" || strings.Contains(sel, "<") || isPlatformSelector(sel) {
			continue
		}
		if scope {
			sel = scopeCSSSelector(sel)
		}
		selectors = append(selectors, sel)
	}
	if len(selectors) == 0 {
		return nil
	}
	decls := filterCSSDeclarations(rule.Declarations)
	if len(decls) == 0 {
		return nil
	}
	rule.Selectors = selectors
	rule.Declarations = decls
	return rule
}

func sanitizeCSSAtRule(rule *css.Rule) *css.Rule {
	if !allowedAtRules[strings.ToLower(strings.TrimSpace(rule.Name))] {
		return nil
	}
	if styleBreakout(rule.Prelude) {
		return nil
	}
	// Keyframe step selectors (from/to/percentages) are not DOM selectors
	// and must not be theme-scoped; only their declarations are filtered.
	scopeNested := strings.ToLower(strings.TrimSpace(rule.Name)) != "@keyframes"
	if rule.EmbedsRules() {
		nested := make([]*css.Rule, 0, len(rule.Rules))
		for _, sub := range rule.Rules {
			if clean := sanitizeCSSRule(sub, scopeNested); clean != nil {
				nested = append(nested, clean)
			}
		}
		if len(nested) == 0 {
			return nil
		}
		rule.Rules = nested
		return rule
	}
	decls := filterCSSDeclarations(rule.Declarations)
	if len(decls) == 0 {
		return nil
	}
	rule.Declarations = decls
	return rule
}

func isPlatformSelector(sel string) bool {
	lower := strings.ToLower(sel)
	for _, frag := range platformSelectorFragments {
		if strings.Contains(lower, frag) {
			return true
		}
	}
	return false
}

func scopeCSSSelector(sel string) string {
	if sel == ":root" {
		return themeScopeRoot
	}
	if strings.HasPrefix(sel, themeScopeRoot) {
		return sel
	}
	return themeScopeRoot + " " + sel
}

func filterCSSDeclarations(decls []*css.Declaration) []*css.Declaration {
	kept := make([]*css.Declaration, 0, len(decls))
	for _, decl := range decls {
		if decl == nil {
			continue
		}
		prop := strings.ToLower(strings.TrimSpace(decl.Property))
		if blockedDeclarationProps[prop] {
			continue
		}
		lowerValue := strings.ToLower(decl.Value)
		blocked := false
		for _, frag := range blockedValueFragments {
			if strings.Contains(lowerValue, frag) {
				blocked = true
				break
			}
		}
		if !blocked {
			for _, target := range cssURLTargets(decl.Value) {
				if !allowedCSSURL(target) {
					blocked = true
					break
				}
			}
		}
		if !blocked {
			kept = append(kept, decl)
		}
	}
	return kept
}

func cssURLTargets(value string) []string {
	matches := cssURLTarget.FindAllStringSubmatch(value, -1)
	targets := make([]string, 0, len(matches))
	for _, m := range matches {
		if len(m) > 1 {
			targets = append(targets, strings.TrimSpace(m[1]))
		}
	}
	return targets
}

// allowedCSSURL permits relative, root-relative, protocol-relative, and
// http(s) targets plus same-document fragments and data:image/* payloads.
// Every other scheme (javascript:, data:text/html, vbscript:, file:, ...) is
// rejected.
func allowedCSSURL(target string) bool {
	t := strings.TrimSpace(target)
	if t == "" || strings.HasPrefix(t, "#") {
		return true
	}
	lower := strings.ToLower(t)
	if strings.HasPrefix(lower, "data:") {
		return strings.HasPrefix(lower, "data:image/")
	}
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		return true
	}
	if idx := strings.Index(lower, ":"); idx >= 0 {
		before := lower[:idx]
		// A colon before any /?# starts a URI scheme; otherwise it is a
		// port or path segment of a relative URL.
		if !strings.ContainsAny(before, "/?#") {
			return false
		}
	}
	return true
}

func styleBreakout(text string) bool {
	return strings.Contains(strings.ToLower(text), "</style")
}

func (s *Service) CheckAdminRole(ctx context.Context, blogID, userID string) error {
	var role string
	err := s.pool.QueryRow(ctx, `SELECT role FROM blog_members WHERE blog_id = $1 AND user_id = $2`, blogID, userID).Scan(&role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &Fault{Status: http.StatusForbidden, Code: "FORBIDDEN", Message: "Not a member of this blog"}
		}
		return err
	}
	if role != "OWNER" && role != "ADMIN" {
		return &Fault{Status: http.StatusForbidden, Code: "INSUFFICIENT_PERMISSIONS", Message: "Requires blog admin privileges"}
	}
	return nil
}

func (s *Service) ListThemes(ctx context.Context) ([]ThemeCatalogItem, error) {
	if s.pool == nil {
		return nil, &Fault{Status: http.StatusServiceUnavailable, Code: "SERVICE_UNAVAILABLE", Message: "Database is unavailable"}
	}

	rows, err := s.pool.Query(ctx, `
		SELECT id, slug, name, description, layout_type, default_tokens_json, preview_image_url, version, is_active, created_at
		FROM themes
		WHERE is_active = TRUE
		ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	themes := make([]ThemeCatalogItem, 0)
	for rows.Next() {
		var t ThemeCatalogItem
		if err := rows.Scan(
			&t.ID, &t.Slug, &t.Name, &t.Description, &t.LayoutType, &t.DefaultTokensJSON,
			&t.PreviewImageURL, &t.Version, &t.IsActive, &t.CreatedAt,
		); err != nil {
			return nil, err
		}
		themes = append(themes, t)
	}
	return themes, rows.Err()
}

func (s *Service) GetBlogTheme(ctx context.Context, blogID string) (*BlogThemeConfig, error) {
	if s.pool == nil {
		return nil, &Fault{Status: http.StatusServiceUnavailable, Code: "SERVICE_UNAVAILABLE", Message: "Database is unavailable"}
	}

	var cfg BlogThemeConfig
	err := s.pool.QueryRow(ctx, `
		SELECT bt.blog_id, bt.theme_id, t.slug, t.name, bt.active_layout,
		       bt.color_palette_tokens, bt.typography_tokens, bt.custom_css_sanitized,
		       bt.dark_mode_preference, bt.updated_at
		FROM blog_themes bt
		JOIN themes t ON t.id = bt.theme_id
		WHERE bt.blog_id = $1
	`, blogID).Scan(
		&cfg.BlogID, &cfg.ThemeID, &cfg.ThemeSlug, &cfg.ThemeName, &cfg.ActiveLayout,
		&cfg.ColorPaletteTokens, &cfg.TypographyTokens, &cfg.CustomCSS,
		&cfg.DarkModePreference, &cfg.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &Fault{Status: http.StatusNotFound, Code: "THEME_NOT_CONFIGURED", Message: "Blog theme configuration not found"}
		}
		return nil, err
	}

	return &cfg, nil
}

func (s *Service) UpdateBlogTheme(ctx context.Context, blogID, userID string, input UpdateBlogThemeInput) (*BlogThemeConfig, error) {
	if s.pool == nil {
		return nil, &Fault{Status: http.StatusServiceUnavailable, Code: "SERVICE_UNAVAILABLE", Message: "Database is unavailable"}
	}
	if err := s.CheckAdminRole(ctx, blogID, userID); err != nil {
		return nil, err
	}

	sanitizedCSS := ""
	if input.CustomCSS != nil {
		sanitizedCSS = SanitizeCSS(*input.CustomCSS)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Fetch current
	var current BlogThemeConfig
	err = tx.QueryRow(ctx, `
		SELECT theme_id, active_layout, color_palette_tokens, typography_tokens, custom_css_sanitized, dark_mode_preference
		FROM blog_themes WHERE blog_id = $1
	`, blogID).Scan(
		&current.ThemeID, &current.ActiveLayout, &current.ColorPaletteTokens,
		&current.TypographyTokens, &current.CustomCSS, &current.DarkModePreference,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &Fault{Status: http.StatusNotFound, Code: "THEME_NOT_FOUND", Message: "Blog theme record not found"}
		}
		return nil, err
	}

	targetThemeID := current.ThemeID
	if input.ThemeID != nil && *input.ThemeID != "" {
		targetThemeID = *input.ThemeID
	}
	targetLayout := current.ActiveLayout
	if input.ActiveLayout != nil && *input.ActiveLayout != "" {
		targetLayout = *input.ActiveLayout
	}
	targetColors := current.ColorPaletteTokens
	if input.ColorPaletteTokens != nil {
		targetColors = *input.ColorPaletteTokens
	}
	targetTypo := current.TypographyTokens
	if input.TypographyTokens != nil {
		targetTypo = *input.TypographyTokens
	}
	targetCSS := current.CustomCSS
	if input.CustomCSS != nil {
		targetCSS = sanitizedCSS
	}
	targetDark := current.DarkModePreference
	if input.DarkModePreference != nil && *input.DarkModePreference != "" {
		targetDark = *input.DarkModePreference
	}

	_, err = tx.Exec(ctx, `
		UPDATE blog_themes
		SET theme_id = $1, active_layout = $2, color_palette_tokens = $3,
		    typography_tokens = $4, custom_css_sanitized = $5,
		    dark_mode_preference = $6, updated_at = NOW()
		WHERE blog_id = $7
	`, targetThemeID, targetLayout, targetColors, targetTypo, targetCSS, targetDark, blogID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return s.GetBlogTheme(ctx, blogID)
}

func (s *Service) ResetBlogTheme(ctx context.Context, blogID, userID string) (*BlogThemeConfig, error) {
	if s.pool == nil {
		return nil, &Fault{Status: http.StatusServiceUnavailable, Code: "SERVICE_UNAVAILABLE", Message: "Database is unavailable"}
	}
	if err := s.CheckAdminRole(ctx, blogID, userID); err != nil {
		return nil, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Fetch default theme
	var defaultThemeID string
	var defaultLayout string
	var defaultTokens json.RawMessage
	err = tx.QueryRow(ctx, `SELECT id, layout_type, default_tokens_json FROM themes WHERE is_active = TRUE ORDER BY (slug = 'forest-editorial') DESC, created_at ASC LIMIT 1`).Scan(&defaultThemeID, &defaultLayout, &defaultTokens)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx, `
		UPDATE blog_themes
		SET theme_id = $1, active_layout = $2, color_palette_tokens = $3,
		    typography_tokens = '{}'::jsonb, custom_css_sanitized = '',
		    dark_mode_preference = 'AUTO', updated_at = NOW()
		WHERE blog_id = $4
	`, defaultThemeID, defaultLayout, defaultTokens, blogID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return s.GetBlogTheme(ctx, blogID)
}
