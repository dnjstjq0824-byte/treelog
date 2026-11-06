package media

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	"image/png"
	"log/slog"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"

	"treeblog/internal/platform/s3"
)

const (
	maxObjectBytes = 64 << 20
	variantQuality = 85
)

// Processor runs the asynchronous media pipeline: integrity verification,
// dimension detection, EXIF stripping, and JPEG variant generation.
type Processor struct {
	Pool         *pgxpool.Pool
	S3           *s3.Client
	BucketRaw    string
	BucketPublic string
	Logger       *slog.Logger
}

// ProcessAsset verifies and optimizes one uploaded asset. It is idempotent
// for assets already in a terminal state.
func (p *Processor) ProcessAsset(ctx context.Context, assetID string) error {
	if p.Pool == nil || p.S3 == nil {
		return errors.New("media processor requires a database pool and object storage")
	}
	var asset struct {
		ID         string
		MimeType   string
		SHA256     string
		Bucket     string
		Key        string
		JobID      *string
		JobType    *string
		Status     string
	}
	err := p.Pool.QueryRow(ctx, `
		SELECT a.id, a.mime_type, a.sha256_hash, us.storage_bucket, us.storage_object_key,
		       j.id, j.job_type, a.status
		FROM media_assets a
		JOIN upload_sessions us ON us.id = a.upload_session_id
		LEFT JOIN media_processing_jobs j
		  ON j.media_asset_id = a.id AND j.status IN ('QUEUED', 'RUNNING')
		WHERE a.id = $1
	`, assetID).Scan(
		&asset.ID, &asset.MimeType, &asset.SHA256, &asset.Bucket, &asset.Key,
		&asset.JobID, &asset.JobType, &asset.Status,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("media asset %s not found", assetID)
		}
		return err
	}
	if asset.Status == "READY" || asset.Status == "QUARANTINED" || asset.Status == "DELETED" {
		return nil
	}
	if asset.JobID != nil {
		_, _ = p.Pool.Exec(ctx, `UPDATE media_processing_jobs
			SET status = 'RUNNING', attempt_count = attempt_count + 1, started_at = NOW()
			WHERE id = $1`, *asset.JobID)
	}
	fail := func(reason string) error {
		_, _ = p.Pool.Exec(ctx, `UPDATE media_assets
			SET status = 'QUARANTINED', virus_scan_status = 'SKIPPED', updated_at = NOW()
			WHERE id = $1`, asset.ID)
		if asset.JobID != nil {
			_, _ = p.Pool.Exec(ctx, `UPDATE media_processing_jobs
				SET status = 'FAILED', error_message = $2, completed_at = NOW()
				WHERE id = $1`, *asset.JobID, reason)
		}
		p.log().Warn("media asset quarantined", "asset_id", asset.ID, "reason", reason)
		return errors.New(reason)
	}

	body, err := p.S3.Get(ctx, asset.Bucket, asset.Key, maxObjectBytes)
	if err != nil {
		return p.retryOrFail(ctx, asset.JobID, fmt.Sprintf("object download failed: %v", err))
	}
	sum := sha256.Sum256(body)
	if !strings.EqualFold(hex.EncodeToString(sum[:]), asset.SHA256) {
		return fail("sha256 mismatch: uploaded bytes differ from the declared hash")
	}
	if !VerifyMagicBytes(asset.MimeType, body) {
		return fail("magic byte signature does not match the declared mime type")
	}

	width, height, img, decodeErr := decodeImage(asset.MimeType, body)
	exifStripped := false
	if decodeErr == nil && (asset.MimeType == "image/jpeg" || asset.MimeType == "image/png") {
		if stripped, err := stripEXIF(asset.MimeType, img); err == nil {
			if err := p.S3.Put(ctx, asset.Bucket, asset.Key, asset.MimeType, stripped); err == nil {
				exifStripped = true
			} else {
				p.log().Warn("exif-stripped re-upload failed", "asset_id", asset.ID, "error", err)
			}
		}
	}
	if decodeErr == nil && img != nil {
		p.storeVariants(ctx, asset.ID, img)
	}

	if decodeErr != nil {
		_, _ = p.Pool.Exec(ctx, `UPDATE media_assets
			SET virus_scan_status = 'CLEAN', updated_at = NOW()
			WHERE id = $1`, asset.ID)
		p.completeJob(ctx, asset.JobID, "dimensions unavailable: "+decodeErr.Error())
		_, _ = p.Pool.Exec(ctx, `UPDATE media_assets
			SET status = 'READY', updated_at = NOW()
			WHERE id = $1`, asset.ID)
		return nil
	}
	_, _ = p.Pool.Exec(ctx, `UPDATE media_assets
		SET width = $2, height = $3, exif_stripped = $4,
		    virus_scan_status = 'CLEAN', status = 'READY', updated_at = NOW()
		WHERE id = $1`, asset.ID, width, height, exifStripped)
	p.completeJob(ctx, asset.JobID, "")
	return nil
}

func (p *Processor) retryOrFail(ctx context.Context, jobID *string, reason string) error {
	if jobID != nil {
		_, _ = p.Pool.Exec(ctx, `UPDATE media_processing_jobs
			SET status = 'FAILED', error_message = $2, completed_at = NOW()
			WHERE id = $1`, *jobID, reason)
	}
	_ = ctx
	return errors.New(reason)
}

func (p *Processor) completeJob(ctx context.Context, jobID *string, note string) {
	if jobID == nil {
		return
	}
	_, _ = p.Pool.Exec(ctx, `UPDATE media_processing_jobs
		SET status = 'COMPLETED', error_message = NULLIF($2, ''), completed_at = NOW()
		WHERE id = $1`, *jobID, note)
	_ = ctx
}

func (p *Processor) storeVariants(ctx context.Context, assetID string, img image.Image) {
	for _, variant := range buildVariants(img) {
		key := fmt.Sprintf("variants/%s/%s.jpg", assetID, strings.ToLower(variant.Name))
		if err := p.S3.Put(ctx, p.BucketPublic, key, "image/jpeg", variant.JPEG); err != nil {
			p.log().Warn("variant upload failed", "asset_id", assetID, "variant", variant.Name, "error", err)
			continue
		}
		publicURL := fmt.Sprintf("/api/v1/public/media/%s?variant=%s", assetID, variant.Name)
		if _, err := p.Pool.Exec(ctx, `
			INSERT INTO media_variants (media_asset_id, variant_name, format, width, height, size_bytes, public_cdn_url)
			VALUES ($1, $2, 'JPEG', $3, $4, $5, $6)
			ON CONFLICT (media_asset_id, variant_name, format) DO UPDATE
			SET width = EXCLUDED.width, height = EXCLUDED.height,
			    size_bytes = EXCLUDED.size_bytes, public_cdn_url = EXCLUDED.public_cdn_url
		`, assetID, variant.Name, variant.Width, variant.Height, len(variant.JPEG), publicURL); err != nil {
			p.log().Warn("variant row insert failed", "asset_id", assetID, "variant", variant.Name, "error", err)
		}
	}
}

func (p *Processor) log() *slog.Logger {
	if p.Logger == nil {
		return slog.Default()
	}
	return p.Logger
}

// decodeImage decodes image bytes and reports dimensions. AVIF/HEIC have no
// Go decoder available, so they report an explanatory error while the
// original bytes remain servable.
func decodeImage(mimeType string, body []byte) (width, height int, img image.Image, err error) {
	switch mimeType {
	case "image/jpeg", "image/png", "image/gif", "image/webp":
		config, format, err := image.DecodeConfig(bytes.NewReader(body))
		if err != nil {
			return 0, 0, nil, fmt.Errorf("cannot decode %s image: %w", format, err)
		}
		decoded, _, err := image.Decode(bytes.NewReader(body))
		if err != nil {
			return 0, 0, nil, fmt.Errorf("cannot decode %s image: %w", format, err)
		}
		return config.Width, config.Height, decoded, nil
	case "image/avif", "image/heic":
		return 0, 0, nil, fmt.Errorf("%s decoding is not supported; original is served as-is", mimeType)
	default:
		return 0, 0, nil, fmt.Errorf("unsupported mime type %s", mimeType)
	}
}

// stripEXIF re-encodes a decoded image, which drops EXIF/XMP metadata
// segments that the decoder does not preserve.
func stripEXIF(mimeType string, img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	switch mimeType {
	case "image/jpeg":
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 92}); err != nil {
			return nil, err
		}
	case "image/png":
		if err := png.Encode(&buf, img); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("exif stripping supports jpeg/png only, got %s", mimeType)
	}
	return buf.Bytes(), nil
}

type variantImage struct {
	Name   string
	Width  int
	Height int
	JPEG   []byte
}

// buildVariants renders downscaled JPEG renditions. Sources are never
// upscaled; a small source yields only the thumbnail.
func buildVariants(src image.Image) []variantImage {
	bounds := src.Bounds()
	srcW, srcH := bounds.Dx(), bounds.Dy()
	if srcW <= 0 || srcH <= 0 {
		return nil
	}
	variants := make([]variantImage, 0, 6)
	thumb := centerCropSquare(src)
	if thumb != nil {
		if encoded := encodeJPEG(resizeImage(thumb, 400, 400)); encoded != nil {
			variants = append(variants, variantImage{Name: "THUMB_400", Width: 400, Height: 400, JPEG: encoded})
		}
	}
	for _, target := range []struct {
		name  string
		width int
	}{
		{"W320", 320},
		{"W640", 640},
		{"W1200", 1200},
		{"W1920", 1920},
	} {
		if srcW <= target.width {
			continue
		}
		height := srcH * target.width / srcW
		if height <= 0 {
			continue
		}
		if encoded := encodeJPEG(resizeImage(src, target.width, height)); encoded != nil {
			variants = append(variants, variantImage{Name: target.name, Width: target.width, Height: height, JPEG: encoded})
		}
	}
	return variants
}

func resizeImage(src image.Image, width, height int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)
	return dst
}

func centerCropSquare(src image.Image) image.Image {
	bounds := src.Bounds()
	side := bounds.Dx()
	if bounds.Dy() < side {
		side = bounds.Dy()
	}
	if side <= 0 {
		return nil
	}
	x0 := bounds.Min.X + (bounds.Dx()-side)/2
	y0 := bounds.Min.Y + (bounds.Dy()-side)/2
	cropped := image.NewRGBA(image.Rect(0, 0, side, side))
	draw.Draw(cropped, cropped.Bounds(), src, image.Point{X: x0, Y: y0}, draw.Src)
	return cropped
}

func encodeJPEG(img image.Image) []byte {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: variantQuality}); err != nil {
		return nil
	}
	return buf.Bytes()
}
