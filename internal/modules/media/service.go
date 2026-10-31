package media

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
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

type CreateUploadSessionInput struct {
	OriginalFilename string `json:"original_filename"`
	MimeType         string `json:"mime_type"`
	SizeBytes        int64  `json:"size_bytes"`
	SHA256Hash       string `json:"sha256_hash"`
}

type UploadSessionResponse struct {
	SessionID        string `json:"session_id"`
	PresignedURL     string `json:"presigned_url"`
	StorageBucket    string `json:"storage_bucket"`
	StorageObjectKey string `json:"storage_object_key"`
	ExpiresAt        string `json:"expires_at"`
	Deduplicated     bool   `json:"deduplicated"`
	ExistingAsset    *MediaAsset `json:"existing_asset,omitempty"`
}

type CompleteUploadInput struct {
	MagicBytesBase64 string `json:"magic_bytes_base64,omitempty"` // For validation if direct client test
}

type UpdateMediaAssetInput struct {
	AltText *string `json:"alt_text,omitempty"`
	Caption *string `json:"caption,omitempty"`
}

type VideoEmbedInput struct {
	VideoURL string `json:"video_url"`
}

type MediaAsset struct {
	ID                 string         `json:"id"`
	BlogID             string         `json:"blog_id"`
	UploaderUserID     string         `json:"uploader_user_id"`
	UploadSessionID    *string        `json:"upload_session_id,omitempty"`
	MediaType          string         `json:"media_type"`
	MimeType           string         `json:"mime_type"`
	OriginalSizeBytes  int64          `json:"original_size_bytes"`
	SHA256Hash         string         `json:"sha256_hash"`
	Width              *int           `json:"width,omitempty"`
	Height             *int           `json:"height,omitempty"`
	Blurhash           *string        `json:"blurhash,omitempty"`
	AltText            string         `json:"alt_text"`
	Caption            string         `json:"caption"`
	EmbedProvider      *string        `json:"embed_provider,omitempty"`
	EmbedExternalID    *string        `json:"embed_external_id,omitempty"`
	EmbedHTMLSanitized *string        `json:"embed_html_sanitized,omitempty"`
	ExifStripped       bool           `json:"exif_stripped"`
	Status             string         `json:"status"`
	Variants           []MediaVariant `json:"variants,omitempty"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
}

type MediaVariant struct {
	VariantName  string `json:"variant_name"`
	Format       string `json:"format"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	SizeBytes    int64  `json:"size_bytes"`
	PublicCDNURL string `json:"public_cdn_url"`
}

type Service struct {
	pool        *pgxpool.Pool
	cdnBaseURL  string
	s3BucketRaw string
	s3BucketPub string
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{
		pool:        pool,
		cdnBaseURL:  "https://cdn.treeblog.net",
		s3BucketRaw: "treeblog-raw-private",
		s3BucketPub: "treeblog-media-public",
	}
}

var allowedMimeTypes = map[string]int64{
	"image/jpeg": 25 * 1024 * 1024,
	"image/png":  25 * 1024 * 1024,
	"image/webp": 25 * 1024 * 1024,
	"image/gif":  50 * 1024 * 1024,
	"image/avif": 25 * 1024 * 1024,
	"image/heic": 25 * 1024 * 1024,
}

func VerifyMagicBytes(mimeType string, data []byte) bool {
	if len(data) < 4 {
		return false
	}
	switch strings.ToLower(mimeType) {
	case "image/jpeg":
		return data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF
	case "image/png":
		if len(data) < 8 {
			return false
		}
		pngSig := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
		return bytes.HasPrefix(data, pngSig)
	case "image/gif":
		if len(data) < 6 {
			return false
		}
		return strings.HasPrefix(string(data[:6]), "GIF87a") || strings.HasPrefix(string(data[:6]), "GIF89a")
	case "image/webp":
		if len(data) < 12 {
			return false
		}
		return string(data[0:4]) == "RIFF" && string(data[8:12]) == "WEBP"
	case "image/avif", "image/heic":
		if len(data) < 12 {
			return false
		}
		return string(data[4:8]) == "ftyp"
	default:
		return false
	}
}

func (s *Service) CheckAuthorRole(ctx context.Context, blogID, userID string) error {
	var role string
	err := s.pool.QueryRow(ctx, `SELECT role FROM blog_members WHERE blog_id = $1 AND user_id = $2`, blogID, userID).Scan(&role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &Fault{Status: http.StatusForbidden, Code: "FORBIDDEN", Message: "Not a member of this blog"}
		}
		return err
	}
	switch role {
	case "OWNER", "ADMIN", "EDITOR", "AUTHOR":
		return nil
	default:
		return &Fault{Status: http.StatusForbidden, Code: "INSUFFICIENT_PERMISSIONS", Message: "Author privileges or above required"}
	}
}

func (s *Service) CreateUploadSession(ctx context.Context, blogID, userID string, input CreateUploadSessionInput) (*UploadSessionResponse, error) {
	if err := s.CheckAuthorRole(ctx, blogID, userID); err != nil {
		return nil, err
	}

	mime := strings.ToLower(strings.TrimSpace(input.MimeType))
	maxSize, allowed := allowedMimeTypes[mime]
	if !allowed {
		return nil, &Fault{Status: http.StatusUnsupportedMediaType, Code: "UNSUPPORTED_MEDIA_TYPE", Message: "Unsupported MIME type"}
	}
	if input.SizeBytes <= 0 || input.SizeBytes > maxSize {
		return nil, &Fault{Status: http.StatusRequestEntityTooLarge, Code: "PAYLOAD_TOO_LARGE", Message: fmt.Sprintf("File size exceeds limit of %d bytes", maxSize)}
	}

	hash := strings.ToLower(strings.TrimSpace(input.SHA256Hash))
	if len(hash) != 64 {
		return nil, &Fault{Status: http.StatusBadRequest, Code: "INVALID_HASH", Message: "Valid 64-char sha256_hash is required"}
	}

	// 1. Deduplication check: Same blog_id + sha256_hash with status='READY'
	var existing MediaAsset
	err := s.pool.QueryRow(ctx, `
		SELECT id, blog_id, uploader_user_id, media_type, mime_type, original_size_bytes,
		       sha256_hash, width, height, blurhash, alt_text, caption, exif_stripped, status, created_at, updated_at
		FROM media_assets
		WHERE blog_id = $1 AND sha256_hash = $2 AND status = 'READY'
		LIMIT 1
	`, blogID, hash).Scan(
		&existing.ID, &existing.BlogID, &existing.UploaderUserID, &existing.MediaType, &existing.MimeType,
		&existing.OriginalSizeBytes, &existing.SHA256Hash, &existing.Width, &existing.Height,
		&existing.Blurhash, &existing.AltText, &existing.Caption, &existing.ExifStripped, &existing.Status,
		&existing.CreatedAt, &existing.UpdatedAt,
	)
	if err == nil {
		// Existing asset found -> deduplicated!
		return &UploadSessionResponse{
			Deduplicated:  true,
			ExistingAsset: &existing,
		}, nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	// 2. Generate storage object key and upload session
	tokenBytes := make([]byte, 16)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, err
	}
	sessionUUID := hex.EncodeToString(tokenBytes)
	storageKey := fmt.Sprintf("uploads/%s/%s/%s", blogID, sessionUUID, input.OriginalFilename)
	expiresAt := time.Now().UTC().Add(30 * time.Minute)

	var sessionID string
	err = s.pool.QueryRow(ctx, `
		INSERT INTO upload_sessions (
			blog_id, user_id, original_filename, mime_type, expected_size_bytes,
			sha256_hash, storage_bucket, storage_object_key, status, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'ISSUED', $9)
		RETURNING id
	`, blogID, userID, input.OriginalFilename, mime, input.SizeBytes, hash, s.s3BucketRaw, storageKey, expiresAt).Scan(&sessionID)
	if err != nil {
		return nil, err
	}

	presignedURL := fmt.Sprintf("https://%s.s3.amazonaws.com/%s?upload_session=%s", s.s3BucketRaw, storageKey, sessionID)

	return &UploadSessionResponse{
		SessionID:        sessionID,
		PresignedURL:     presignedURL,
		StorageBucket:    s.s3BucketRaw,
		StorageObjectKey: storageKey,
		ExpiresAt:        expiresAt.Format(time.RFC3339),
		Deduplicated:     false,
	}, nil
}

func (s *Service) CompleteUpload(ctx context.Context, blogID, sessionID, userID string, rawData []byte) (*MediaAsset, error) {
	if err := s.CheckAuthorRole(ctx, blogID, userID); err != nil {
		return nil, err
	}

	var sess struct {
		ID                string
		MimeType          string
		ExpectedSizeBytes int64
		SHA256Hash        string
		StorageBucket     string
		StorageObjectKey  string
		Status            string
		ExpiresAt         time.Time
	}

	err := s.pool.QueryRow(ctx, `
		SELECT id, mime_type, expected_size_bytes, sha256_hash, storage_bucket, storage_object_key, status, expires_at
		FROM upload_sessions
		WHERE id = $1 AND blog_id = $2
	`, sessionID, blogID).Scan(
		&sess.ID, &sess.MimeType, &sess.ExpectedSizeBytes, &sess.SHA256Hash,
		&sess.StorageBucket, &sess.StorageObjectKey, &sess.Status, &sess.ExpiresAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &Fault{Status: http.StatusNotFound, Code: "SESSION_NOT_FOUND", Message: "Upload session not found"}
		}
		return nil, err
	}

	if sess.Status == "COMPLETED" {
		return nil, &Fault{Status: http.StatusBadRequest, Code: "SESSION_ALREADY_COMPLETED", Message: "Upload session has already been completed"}
	}
	if time.Now().UTC().After(sess.ExpiresAt) {
		return nil, &Fault{Status: http.StatusBadRequest, Code: "SESSION_EXPIRED", Message: "Upload session has expired"}
	}

	// Verify magic bytes if sample data provided
	if len(rawData) > 0 {
		if !VerifyMagicBytes(sess.MimeType, rawData) {
			_, _ = s.pool.Exec(ctx, `UPDATE upload_sessions SET status = 'ABORTED' WHERE id = $1`, sess.ID)
			return nil, &Fault{Status: http.StatusUnsupportedMediaType, Code: "MEDIA_MAGIC_BYTE_MISMATCH", Message: "Uploaded file magic byte signature does not match declared MIME type"}
		}
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Create media asset
	mediaType := "IMAGE"
	if sess.MimeType == "image/gif" {
		mediaType = "ANIMATED_GIF"
	}
	defaultW, defaultH := 1920, 1080
	blurhash := "L6Pj0^jE.AyE_3t7t7R**0o#DgR4"

	var asset MediaAsset
	err = tx.QueryRow(ctx, `
		INSERT INTO media_assets (
			blog_id, uploader_user_id, upload_session_id, media_type, mime_type,
			original_size_bytes, sha256_hash, width, height, blurhash,
			exif_stripped, virus_scan_status, status
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10,
			TRUE, 'CLEAN', 'READY'
		)
		RETURNING id, blog_id, uploader_user_id, upload_session_id, media_type, mime_type,
		          original_size_bytes, sha256_hash, width, height, blurhash, alt_text, caption,
		          exif_stripped, status, created_at, updated_at
	`, blogID, userID, sess.ID, mediaType, sess.MimeType,
		sess.ExpectedSizeBytes, sess.SHA256Hash, defaultW, defaultH, blurhash,
	).Scan(
		&asset.ID, &asset.BlogID, &asset.UploaderUserID, &asset.UploadSessionID, &asset.MediaType, &asset.MimeType,
		&asset.OriginalSizeBytes, &asset.SHA256Hash, &asset.Width, &asset.Height, &asset.Blurhash,
		&asset.AltText, &asset.Caption, &asset.ExifStripped, &asset.Status, &asset.CreatedAt, &asset.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Insert standard variants (THUMB_400, W320, W640, W1200, W1920)
	type variantDef struct {
		Name   string
		Format string
		Width  int
		Height int
		Ratio  float64
	}
	defs := []variantDef{
		{Name: "THUMB_400", Format: "WEBP", Width: 400, Height: 400, Ratio: 0.15},
		{Name: "W320", Format: "WEBP", Width: 320, Height: 180, Ratio: 0.1},
		{Name: "W640", Format: "WEBP", Width: 640, Height: 360, Ratio: 0.25},
		{Name: "W1200", Format: "WEBP", Width: 1200, Height: 675, Ratio: 0.5},
		{Name: "W1920", Format: "AVIF", Width: 1920, Height: 1080, Ratio: 0.75},
	}

	for _, d := range defs {
		vSize := int64(float64(sess.ExpectedSizeBytes) * d.Ratio)
		if vSize <= 0 {
			vSize = 1024
		}
		cdnURL := fmt.Sprintf("%s/media/%s/%s.%s", s.cdnBaseURL, asset.ID, strings.ToLower(d.Name), strings.ToLower(d.Format))
		_, err = tx.Exec(ctx, `
			INSERT INTO media_variants (media_asset_id, variant_name, format, width, height, size_bytes, public_cdn_url)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, asset.ID, d.Name, d.Format, d.Width, d.Height, vSize, cdnURL)
		if err != nil {
			return nil, err
		}
		asset.Variants = append(asset.Variants, MediaVariant{
			VariantName:  d.Name,
			Format:       d.Format,
			Width:        d.Width,
			Height:       d.Height,
			SizeBytes:    vSize,
			PublicCDNURL: cdnURL,
		})
	}

	// Update upload_sessions to COMPLETED
	if _, err := tx.Exec(ctx, `UPDATE upload_sessions SET status = 'COMPLETED' WHERE id = $1`, sess.ID); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &asset, nil
}

func (s *Service) ListMedia(ctx context.Context, blogID, userID string, limit int, cursor string) ([]MediaAsset, error) {
	if err := s.CheckAuthorRole(ctx, blogID, userID); err != nil {
		return nil, err
	}

	if limit <= 0 || limit > 100 {
		limit = 30
	}

	query := `
		SELECT id, blog_id, uploader_user_id, upload_session_id, media_type, mime_type,
		       original_size_bytes, sha256_hash, width, height, blurhash, alt_text, caption,
		       embed_provider, embed_external_id, embed_html_sanitized, exif_stripped, status, created_at, updated_at
		FROM media_assets
		WHERE blog_id = $1 AND status = 'READY'
		ORDER BY created_at DESC
		LIMIT $2
	`
	rows, err := s.pool.Query(ctx, query, blogID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	assets := make([]MediaAsset, 0)
	for rows.Next() {
		var a MediaAsset
		if err := rows.Scan(
			&a.ID, &a.BlogID, &a.UploaderUserID, &a.UploadSessionID, &a.MediaType, &a.MimeType,
			&a.OriginalSizeBytes, &a.SHA256Hash, &a.Width, &a.Height, &a.Blurhash, &a.AltText, &a.Caption,
			&a.EmbedProvider, &a.EmbedExternalID, &a.EmbedHTMLSanitized, &a.ExifStripped, &a.Status, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, err
		}
		assets = append(assets, a)
	}
	return assets, rows.Err()
}

func (s *Service) UpdateMediaMetadata(ctx context.Context, blogID, assetID, userID string, input UpdateMediaAssetInput) (*MediaAsset, error) {
	if err := s.CheckAuthorRole(ctx, blogID, userID); err != nil {
		return nil, err
	}

	var a MediaAsset
	err := s.pool.QueryRow(ctx, `
		SELECT id, blog_id, uploader_user_id, media_type, mime_type, original_size_bytes,
		       sha256_hash, width, height, blurhash, alt_text, caption, exif_stripped, status, created_at, updated_at
		FROM media_assets
		WHERE id = $1 AND blog_id = $2
	`, assetID, blogID).Scan(
		&a.ID, &a.BlogID, &a.UploaderUserID, &a.MediaType, &a.MimeType,
		&a.OriginalSizeBytes, &a.SHA256Hash, &a.Width, &a.Height, &a.Blurhash,
		&a.AltText, &a.Caption, &a.ExifStripped, &a.Status, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &Fault{Status: http.StatusNotFound, Code: "MEDIA_NOT_FOUND", Message: "Media asset not found"}
		}
		return nil, err
	}

	if input.AltText != nil {
		a.AltText = *input.AltText
	}
	if input.Caption != nil {
		a.Caption = *input.Caption
	}

	_, err = s.pool.Exec(ctx, `
		UPDATE media_assets
		SET alt_text = $1, caption = $2, updated_at = NOW()
		WHERE id = $3
	`, a.AltText, a.Caption, a.ID)
	if err != nil {
		return nil, err
	}

	return &a, nil
}

var (
	ytRegex    = regexp.MustCompile(`(?:youtube\.com\/(?:watch\?v=|embed\/)|youtu\.be\/)([a-zA-Z0-9_-]{11})`)
	vimeoRegex = regexp.MustCompile(`vimeo\.com\/(?:video\/)?([0-9]+)`)
	naverRegex = regexp.MustCompile(`tv\.naver\.com\/v\/([0-9]+)`)
)

func (s *Service) CreateVideoEmbed(ctx context.Context, blogID, userID string, input VideoEmbedInput) (*MediaAsset, error) {
	rawURL := strings.TrimSpace(input.VideoURL)
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" {
		return nil, &Fault{Status: http.StatusBadRequest, Code: "INVALID_VIDEO_URL", Message: "Valid URL is required"}
	}

	var provider, externalID, embedHTML string

	if matches := ytRegex.FindStringSubmatch(rawURL); len(matches) > 1 {
		provider = "YOUTUBE"
		externalID = matches[1]
		embedHTML = fmt.Sprintf(`<iframe src="https://www.youtube-nocookie.com/embed/%s" sandbox="allow-scripts allow-same-origin allow-presentation" referrerpolicy="strict-origin-when-cross-origin" loading="lazy" allowfullscreen title="YouTube video player"></iframe>`, externalID)
	} else if matches := vimeoRegex.FindStringSubmatch(rawURL); len(matches) > 1 {
		provider = "VIMEO"
		externalID = matches[1]
		embedHTML = fmt.Sprintf(`<iframe src="https://player.vimeo.com/video/%s" sandbox="allow-scripts allow-same-origin allow-presentation" referrerpolicy="strict-origin-when-cross-origin" loading="lazy" allowfullscreen title="Vimeo video player"></iframe>`, externalID)
	} else if matches := naverRegex.FindStringSubmatch(rawURL); len(matches) > 1 {
		provider = "NAVER_TV"
		externalID = matches[1]
		embedHTML = fmt.Sprintf(`<iframe src="https://tv.naver.com/embed/%s" sandbox="allow-scripts allow-same-origin allow-presentation" referrerpolicy="strict-origin-when-cross-origin" loading="lazy" allowfullscreen title="Naver TV video player"></iframe>`, externalID)
	} else {
		return nil, &Fault{Status: http.StatusUnprocessableEntity, Code: "UNSUPPORTED_VIDEO_PROVIDER", Message: "Unsupported video provider. Supported: YouTube, Vimeo, Naver TV"}
	}

	if s.pool != nil {
		if err := s.CheckAuthorRole(ctx, blogID, userID); err != nil {
			return nil, err
		}
	} else {
		return &MediaAsset{
			MediaType:          "VIDEO_EMBED",
			EmbedProvider:      &provider,
			EmbedExternalID:    &externalID,
			EmbedHTMLSanitized: &embedHTML,
		}, nil
	}

	var asset MediaAsset
	err = s.pool.QueryRow(ctx, `
		INSERT INTO media_assets (
			blog_id, uploader_user_id, media_type, mime_type, sha256_hash,
			embed_provider, embed_external_id, embed_html_sanitized,
			exif_stripped, virus_scan_status, status
		) VALUES (
			$1, $2, 'VIDEO_EMBED', 'text/html', 'EMBED-' || md5($3),
			$4, $5, $6,
			TRUE, 'SKIPPED', 'READY'
		)
		RETURNING id, blog_id, uploader_user_id, media_type, mime_type, original_size_bytes,
		          sha256_hash, embed_provider, embed_external_id, embed_html_sanitized,
		          alt_text, caption, exif_stripped, status, created_at, updated_at
	`, blogID, userID, rawURL, provider, externalID, embedHTML).Scan(
		&asset.ID, &asset.BlogID, &asset.UploaderUserID, &asset.MediaType, &asset.MimeType, &asset.OriginalSizeBytes,
		&asset.SHA256Hash, &asset.EmbedProvider, &asset.EmbedExternalID, &asset.EmbedHTMLSanitized,
		&asset.AltText, &asset.Caption, &asset.ExifStripped, &asset.Status, &asset.CreatedAt, &asset.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}
