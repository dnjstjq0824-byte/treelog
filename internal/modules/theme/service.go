package theme

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"time"

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

var dangerousCSSPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)javascript:`),
	regexp.MustCompile(`(?i)expression\s*\(`),
	regexp.MustCompile(`(?i)behavior\s*:`),
	regexp.MustCompile(`(?i)@import`),
	regexp.MustCompile(`(?i)url\s*\(\s*['"]?data:[^'"]*['"]?\s*\)`),
	regexp.MustCompile(`(?i)<[^>]*>`),
	regexp.MustCompile(`(?i)-moz-binding`),
}

func SanitizeCSS(rawCSS string) string {
	cleaned := rawCSS
	for _, pattern := range dangerousCSSPatterns {
		cleaned = pattern.ReplaceAllString(cleaned, "/* stripped unsafe css */")
	}
	return cleaned
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
		return []ThemeCatalogItem{
			{
				ID:          "mock-classic-theme",
				Slug:        "classic-blog",
				Name:        "Classic TreeBlog",
				Description: "Official clean classic blog theme",
				LayoutType:  "CLASSIC_BLOG",
				Version:     "1.0.0",
				IsActive:    true,
			},
		}, nil
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
		return &BlogThemeConfig{
			BlogID:             blogID,
			ThemeID:            "mock-classic-theme",
			ThemeSlug:          "classic-blog",
			ThemeName:          "Classic TreeBlog",
			ActiveLayout:       "CLASSIC_BLOG",
			ColorPaletteTokens: json.RawMessage(`{}`),
			TypographyTokens:   json.RawMessage(`{}`),
			DarkModePreference: "AUTO",
		}, nil
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
	if s.pool != nil {
		if err := s.CheckAdminRole(ctx, blogID, userID); err != nil {
			return nil, err
		}
	}

	sanitizedCSS := ""
	if input.CustomCSS != nil {
		sanitizedCSS = SanitizeCSS(*input.CustomCSS)
	}

	if s.pool == nil {
		return &BlogThemeConfig{
			BlogID:    blogID,
			CustomCSS: sanitizedCSS,
		}, nil
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
	if s.pool != nil {
		if err := s.CheckAdminRole(ctx, blogID, userID); err != nil {
			return nil, err
		}
	}
	if s.pool == nil {
		return &BlogThemeConfig{BlogID: blogID}, nil
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
	err = tx.QueryRow(ctx, `SELECT id, layout_type, default_tokens_json FROM themes WHERE slug = 'classic-blog' OR is_active = TRUE LIMIT 1`).Scan(&defaultThemeID, &defaultLayout, &defaultTokens)
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
