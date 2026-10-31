package blog

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Fault struct {
	Status  int
	Code    string
	Message string
}

func (f *Fault) Error() string { return f.Code }

type CreateBlogInput struct {
	Handle          string `json:"handle"`
	Title           string `json:"title"`
	Subtitle        string `json:"subtitle"`
	Description     string `json:"description"`
	BlogType        string `json:"blog_type"`
	DefaultLanguage string `json:"default_language"`
	Timezone        string `json:"timezone"`
}

type BlogInfo struct {
	ID              string    `json:"id"`
	OwnerUserID     string    `json:"owner_user_id"`
	Handle          string    `json:"handle"`
	BlogType        string    `json:"blog_type"`
	Title           string    `json:"title"`
	Subtitle        string    `json:"subtitle"`
	Description     string    `json:"description"`
	DefaultLanguage string    `json:"default_language"`
	Timezone        string    `json:"timezone"`
	PrimaryURLType  string    `json:"primary_url_type"`
	Status          string    `json:"status"`
	SubscriberCount int64     `json:"subscriber_count"`
	NeighborCount   int64     `json:"neighbor_count"`
	PostCount       int       `json:"post_count"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type BlogSettings struct {
	BlogID                    string    `json:"blog_id"`
	DefaultPostVisibility     string    `json:"default_post_visibility"`
	DefaultAllowComments      string    `json:"default_allow_comments"`
	AllowGuestbook            bool      `json:"allow_guestbook"`
	AllowMutualRequests       bool      `json:"allow_mutual_requests"`
	AllowRSSFullContent       bool      `json:"allow_rss_full_content"`
	AllowSearchIndexing       bool      `json:"allow_search_indexing"`
	AllowTreeScanSync         bool      `json:"allow_treescan_sync"`
	AntiCopyProtection        bool      `json:"anti_copy_protection"`
	RequireApprovalForAuthors bool      `json:"require_approval_for_authors"`
	PostsPerPage              int       `json:"posts_per_page"`
	ShowVisitorCounter        bool      `json:"show_visitor_counter"`
	AdultContentWarning       bool      `json:"adult_content_warning"`
	UpdatedAt                 time.Time `json:"updated_at"`
}

type BlogMemberInfo struct {
	ID           string    `json:"id"`
	BlogID       string    `json:"blog_id"`
	UserID       string    `json:"user_id"`
	Role         string    `json:"role"`
	InviteStatus string    `json:"invite_status"`
	JoinedAt     time.Time `json:"joined_at"`
	Username     string    `json:"username,omitempty"`
	DisplayName  string    `json:"display_name,omitempty"`
}

type GuestbookEntry struct {
	ID              string    `json:"id"`
	BlogID          string    `json:"blog_id"`
	AuthorUserID    string    `json:"author_user_id"`
	ParentEntryID   *string   `json:"parent_entry_id"`
	ContentText     string    `json:"content_text"`
	IsSecret        bool      `json:"is_secret"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	AuthorUsername  string    `json:"author_username,omitempty"`
	AuthorName      string    `json:"author_name,omitempty"`
}

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

func (s *Service) CheckHandleAvailable(ctx context.Context, handle string) (bool, string, error) {
	valid, reason := ValidateHandle(handle)
	if !valid {
		return false, reason, nil
	}
	handle = strings.ToLower(strings.TrimSpace(handle))
	var exists bool
	err := s.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM blogs WHERE handle=$1)", handle).Scan(&exists)
	if err != nil {
		return false, "", err
	}
	if exists {
		return false, "Handle is already taken", nil
	}
	return true, "Handle is available", nil
}

func (s *Service) CreateBlog(ctx context.Context, ownerUserID string, input CreateBlogInput) (*BlogInfo, error) {
	valid, reason := ValidateHandle(input.Handle)
	if !valid {
		return nil, &Fault{Status: 400, Code: "INVALID_REQUEST_PAYLOAD", Message: reason}
	}
	handle := strings.ToLower(strings.TrimSpace(input.Handle))
	title := strings.TrimSpace(input.Title)
	if len(title) < 1 || len(title) > 120 {
		return nil, &Fault{Status: 400, Code: "INVALID_REQUEST_PAYLOAD", Message: "Title must be between 1 and 120 characters"}
	}

	blogType := strings.ToUpper(strings.TrimSpace(input.BlogType))
	if blogType == "" {
		blogType = "PERSONAL"
	}
	if blogType != "PERSONAL" && blogType != "PROFESSIONAL" && blogType != "CORPORATE" && blogType != "COMMUNITY" {
		return nil, &Fault{Status: 400, Code: "INVALID_REQUEST_PAYLOAD", Message: "Invalid blog type"}
	}

	lang := input.DefaultLanguage
	if lang == "" {
		lang = "ko-KR"
	}
	tz := input.Timezone
	if tz == "" {
		tz = "Asia/Seoul"
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	blogID := uuid.NewString()

	// 1. Insert into blogs
	var b BlogInfo
	err = tx.QueryRow(ctx, `INSERT INTO blogs (
		id, owner_user_id, handle, blog_type, title, subtitle, description,
		default_language, timezone, primary_url_type, status
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7,
		$8, $9, 'SUBDOMAIN', 'ACTIVE'
	) RETURNING id, owner_user_id, handle, blog_type, title, subtitle, description,
		default_language, timezone, primary_url_type, status,
		subscriber_count, neighbor_count, post_count, created_at, updated_at`,
		blogID, ownerUserID, handle, blogType, title, strings.TrimSpace(input.Subtitle), strings.TrimSpace(input.Description),
		lang, tz,
	).Scan(
		&b.ID, &b.OwnerUserID, &b.Handle, &b.BlogType, &b.Title, &b.Subtitle, &b.Description,
		&b.DefaultLanguage, &b.Timezone, &b.PrimaryURLType, &b.Status,
		&b.SubscriberCount, &b.NeighborCount, &b.PostCount, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, &Fault{Status: 409, Code: "DUPLICATE_EMAIL_OR_HANDLE", Message: "Handle is already taken"}
		}
		return nil, err
	}

	// 2. Insert into blog_members as OWNER
	if _, err := tx.Exec(ctx, `INSERT INTO blog_members (
		blog_id, user_id, role, invite_status
	) VALUES ($1, $2, 'OWNER', 'ACCEPTED')`, blogID, ownerUserID); err != nil {
		return nil, err
	}

	// 3. Insert default blog_settings
	if _, err := tx.Exec(ctx, `INSERT INTO blog_settings (
		blog_id, default_post_visibility, default_allow_comments,
		allow_guestbook, allow_mutual_requests, allow_rss_full_content,
		allow_search_indexing, allow_treescan_sync, anti_copy_protection,
		require_approval_for_authors, posts_per_page, show_visitor_counter,
		adult_content_warning
	) VALUES (
		$1, 'PUBLIC', 'ALL_ALLOWED',
		TRUE, TRUE, TRUE,
		TRUE, TRUE, FALSE,
		FALSE, 10, TRUE,
		FALSE
	)`, blogID); err != nil {
		return nil, err
	}

	// 4. Insert default subdomain in blog_domains
	subdomain := fmt.Sprintf("%s.treeblog.local", handle)
	verificationToken := uuid.NewString()
	if _, err := tx.Exec(ctx, `INSERT INTO blog_domains (
		blog_id, domain_name, domain_type, is_primary,
		verification_token, verification_status, tls_status
	) VALUES (
		$1, $2, 'SUBDOMAIN', TRUE,
		$3, 'VERIFIED', 'ACTIVE'
	)`, blogID, subdomain, verificationToken); err != nil {
		return nil, err
	}

	// 5. Insert default blog_themes (forest-editorial if exists, or fallback)
	var defaultThemeID string
	err = tx.QueryRow(ctx, "SELECT id FROM themes WHERE slug='forest-editorial' LIMIT 1").Scan(&defaultThemeID)
	if err != nil {
		// If seed not loaded, pick any active theme
		_ = tx.QueryRow(ctx, "SELECT id FROM themes LIMIT 1").Scan(&defaultThemeID)
	}
	if defaultThemeID != "" {
		if _, err := tx.Exec(ctx, `INSERT INTO blog_themes (
			blog_id, theme_id, color_mode
		) VALUES ($1, $2, 'SYSTEM')`, blogID, defaultThemeID); err != nil {
			return nil, err
		}
	}

	// 6. Insert default "미분류" (Uncategorized) category
	if _, err := tx.Exec(ctx, `INSERT INTO blog_categories (
		blog_id, name, slug, description, sort_order
	) VALUES (
		$1, '미분류', 'uncategorized', '기본 카테고리입니다.', 0
	)`, blogID); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &b, nil
}

func (s *Service) GetBlog(ctx context.Context, blogIDOrHandle string) (*BlogInfo, error) {
	var b BlogInfo
	query := `SELECT id, owner_user_id, handle, blog_type, title, subtitle, description,
		default_language, timezone, primary_url_type, status,
		subscriber_count, neighbor_count, post_count, created_at, updated_at
		FROM blogs WHERE `
	if _, err := uuid.Parse(blogIDOrHandle); err == nil {
		query += "id = $1"
	} else {
		query += "handle = $1"
	}

	err := s.db.QueryRow(ctx, query, blogIDOrHandle).Scan(
		&b.ID, &b.OwnerUserID, &b.Handle, &b.BlogType, &b.Title, &b.Subtitle, &b.Description,
		&b.DefaultLanguage, &b.Timezone, &b.PrimaryURLType, &b.Status,
		&b.SubscriberCount, &b.NeighborCount, &b.PostCount, &b.CreatedAt, &b.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, &Fault{Status: 404, Code: "RESOURCE_NOT_FOUND", Message: "Blog not found"}
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (s *Service) ListMyBlogs(ctx context.Context, userID string) ([]BlogInfo, error) {
	rows, err := s.db.Query(ctx, `SELECT b.id, b.owner_user_id, b.handle, b.blog_type, b.title, b.subtitle, b.description,
		b.default_language, b.timezone, b.primary_url_type, b.status,
		b.subscriber_count, b.neighbor_count, b.post_count, b.created_at, b.updated_at
		FROM blogs b
		JOIN blog_members bm ON bm.blog_id = b.id
		WHERE bm.user_id = $1 AND bm.invite_status = 'ACCEPTED' AND b.status != 'CLOSED'
		ORDER BY b.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blogs []BlogInfo
	for rows.Next() {
		var b BlogInfo
		if err := rows.Scan(
			&b.ID, &b.OwnerUserID, &b.Handle, &b.BlogType, &b.Title, &b.Subtitle, &b.Description,
			&b.DefaultLanguage, &b.Timezone, &b.PrimaryURLType, &b.Status,
			&b.SubscriberCount, &b.NeighborCount, &b.PostCount, &b.CreatedAt, &b.UpdatedAt,
		); err != nil {
			return nil, err
		}
		blogs = append(blogs, b)
	}
	return blogs, nil
}

func (s *Service) GetBlogSettings(ctx context.Context, blogID string) (*BlogSettings, error) {
	var st BlogSettings
	err := s.db.QueryRow(ctx, `SELECT blog_id, default_post_visibility, default_allow_comments,
		allow_guestbook, allow_mutual_requests, allow_rss_full_content,
		allow_search_indexing, allow_treescan_sync, anti_copy_protection,
		require_approval_for_authors, posts_per_page, show_visitor_counter,
		adult_content_warning, updated_at
		FROM blog_settings WHERE blog_id=$1`, blogID).Scan(
		&st.BlogID, &st.DefaultPostVisibility, &st.DefaultAllowComments,
		&st.AllowGuestbook, &st.AllowMutualRequests, &st.AllowRSSFullContent,
		&st.AllowSearchIndexing, &st.AllowTreeScanSync, &st.AntiCopyProtection,
		&st.RequireApprovalForAuthors, &st.PostsPerPage, &st.ShowVisitorCounter,
		&st.AdultContentWarning, &st.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, &Fault{Status: 404, Code: "RESOURCE_NOT_FOUND", Message: "Blog settings not found"}
	}
	if err != nil {
		return nil, err
	}
	return &st, nil
}

func (s *Service) UpdateBlogSettings(ctx context.Context, blogID string, st BlogSettings) error {
	if st.PostsPerPage < 5 || st.PostsPerPage > 50 {
		return &Fault{Status: 400, Code: "INVALID_REQUEST_PAYLOAD", Message: "Posts per page must be between 5 and 50"}
	}
	_, err := s.db.Exec(ctx, `UPDATE blog_settings SET
		default_post_visibility = $2,
		default_allow_comments = $3,
		allow_guestbook = $4,
		allow_mutual_requests = $5,
		allow_rss_full_content = $6,
		allow_search_indexing = $7,
		allow_treescan_sync = $8,
		anti_copy_protection = $9,
		require_approval_for_authors = $10,
		posts_per_page = $11,
		show_visitor_counter = $12,
		adult_content_warning = $13,
		updated_at = NOW()
		WHERE blog_id = $1`,
		blogID, st.DefaultPostVisibility, st.DefaultAllowComments,
		st.AllowGuestbook, st.AllowMutualRequests, st.AllowRSSFullContent,
		st.AllowSearchIndexing, st.AllowTreeScanSync, st.AntiCopyProtection,
		st.RequireApprovalForAuthors, st.PostsPerPage, st.ShowVisitorCounter,
		st.AdultContentWarning,
	)
	return err
}

func (s *Service) ListMembers(ctx context.Context, blogID string) ([]BlogMemberInfo, error) {
	rows, err := s.db.Query(ctx, `SELECT bm.id, bm.blog_id, bm.user_id, bm.role, bm.invite_status, bm.joined_at,
		p.username, p.display_name
		FROM blog_members bm
		JOIN user_profiles p ON p.user_id = bm.user_id
		WHERE bm.blog_id = $1
		ORDER BY bm.joined_at ASC`, blogID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []BlogMemberInfo
	for rows.Next() {
		var item BlogMemberInfo
		if err := rows.Scan(&item.ID, &item.BlogID, &item.UserID, &item.Role, &item.InviteStatus, &item.JoinedAt, &item.Username, &item.DisplayName); err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return list, nil
}

func (s *Service) InviteMember(ctx context.Context, blogID, inviterUserID, targetEmail, role string) (*BlogMemberInfo, error) {
	role = strings.ToUpper(strings.TrimSpace(role))
	if role != "ADMIN" && role != "EDITOR" && role != "AUTHOR" && role != "ANALYST" {
		return nil, &Fault{Status: 400, Code: "INVALID_REQUEST_PAYLOAD", Message: "Invalid member role"}
	}

	var targetUserID string
	err := s.db.QueryRow(ctx, "SELECT id FROM users WHERE email=$1", strings.ToLower(strings.TrimSpace(targetEmail))).Scan(&targetUserID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, &Fault{Status: 404, Code: "USER_NOT_FOUND", Message: "User with this email not found"}
	}
	if err != nil {
		return nil, err
	}

	var member BlogMemberInfo
	member.BlogID = blogID
	member.UserID = targetUserID
	member.Role = role
	member.InviteStatus = "ACCEPTED" // Auto accept for now (or pending)
	member.JoinedAt = time.Now()

	err = s.db.QueryRow(ctx, `INSERT INTO blog_members (
		blog_id, user_id, role, invite_status, invited_by, joined_at
	) VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT (blog_id, user_id) DO UPDATE SET role = EXCLUDED.role
	RETURNING id, joined_at`,
		blogID, targetUserID, role, member.InviteStatus, inviterUserID, member.JoinedAt,
	).Scan(&member.ID, &member.JoinedAt)
	if err != nil {
		return nil, err
	}

	return &member, nil
}

func (s *Service) RemoveMember(ctx context.Context, blogID, targetUserID string) error {
	var role string
	err := s.db.QueryRow(ctx, "SELECT role FROM blog_members WHERE blog_id=$1 AND user_id=$2", blogID, targetUserID).Scan(&role)
	if errors.Is(err, pgx.ErrNoRows) {
		return &Fault{Status: 404, Code: "RESOURCE_NOT_FOUND", Message: "Member not found"}
	}
	if role == "OWNER" {
		return &Fault{Status: 403, Code: "CANNOT_REMOVE_OWNER", Message: "Blog owner cannot be removed"}
	}

	_, err = s.db.Exec(ctx, "DELETE FROM blog_members WHERE blog_id=$1 AND user_id=$2", blogID, targetUserID)
	return err
}

func (s *Service) CreateGuestbookEntry(ctx context.Context, blogID, authorUserID string, contentText string, isSecret bool, parentEntryID *string) (*GuestbookEntry, error) {
	contentText = strings.TrimSpace(contentText)
	if len(contentText) < 1 || len(contentText) > 2000 {
		return nil, &Fault{Status: 400, Code: "INVALID_REQUEST_PAYLOAD", Message: "Guestbook content must be between 1 and 2000 characters"}
	}

	var entry GuestbookEntry
	entry.BlogID = blogID
	entry.AuthorUserID = authorUserID
	entry.ParentEntryID = parentEntryID
	entry.ContentText = contentText
	entry.IsSecret = isSecret
	entry.Status = "PUBLISHED"

	err := s.db.QueryRow(ctx, `INSERT INTO guestbook_entries (
		blog_id, author_user_id, parent_entry_id, content_text, is_secret, status
	) VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id, created_at`,
		blogID, authorUserID, parentEntryID, contentText, isSecret, entry.Status,
	).Scan(&entry.ID, &entry.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func (s *Service) ListGuestbookEntries(ctx context.Context, blogID string, viewerUserID string, isBlogStaff bool) ([]GuestbookEntry, error) {
	rows, err := s.db.Query(ctx, `SELECT g.id, g.blog_id, g.author_user_id, g.parent_entry_id, g.content_text,
		g.is_secret, g.status, g.created_at, p.username, p.display_name
		FROM guestbook_entries g
		JOIN user_profiles p ON p.user_id = g.author_user_id
		WHERE g.blog_id = $1 AND g.status = 'PUBLISHED'
		ORDER BY g.created_at DESC LIMIT 100`, blogID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []GuestbookEntry
	for rows.Next() {
		var e GuestbookEntry
		if err := rows.Scan(
			&e.ID, &e.BlogID, &e.AuthorUserID, &e.ParentEntryID, &e.ContentText,
			&e.IsSecret, &e.Status, &e.CreatedAt, &e.AuthorUsername, &e.AuthorName,
		); err != nil {
			return nil, err
		}
		// Server-side masking for secret guestbook entries:
		// only author and blog staff (ADMIN/OWNER/EDITOR) can view
		if e.IsSecret && !isBlogStaff && e.AuthorUserID != viewerUserID {
			e.ContentText = "비밀 방명록 글입니다. 작성자와 블로그 운영자만 확인할 수 있습니다."
		}
		entries = append(entries, e)
	}
	return entries, nil
}

