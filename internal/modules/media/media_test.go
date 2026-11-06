package media

import (
	"context"
	"testing"
	"time"
)

func TestVerifyMagicBytes(t *testing.T) {
	jpegHeader := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10}
	if !VerifyMagicBytes("image/jpeg", jpegHeader) {
		t.Errorf("expected JPEG magic bytes to pass")
	}

	pngHeader := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	if !VerifyMagicBytes("image/png", pngHeader) {
		t.Errorf("expected PNG magic bytes to pass")
	}

	gifHeader := []byte("GIF89a\x01\x00")
	if !VerifyMagicBytes("image/gif", gifHeader) {
		t.Errorf("expected GIF magic bytes to pass")
	}

	webpHeader := []byte("RIFF\x00\x00\x00\x00WEBPVP8 ")
	if !VerifyMagicBytes("image/webp", webpHeader) {
		t.Errorf("expected WebP magic bytes to pass")
	}

	// Mismatched
	if VerifyMagicBytes("image/jpeg", pngHeader) {
		t.Errorf("expected PNG bytes to fail for JPEG")
	}
}

func TestVideoEmbedParsing(t *testing.T) {
	svc := &Service{}

	tests := []struct {
		url      string
		expected bool
	}{
		{"https://www.youtube.com/watch?v=dQw4w9WgXcQ", true},
		{"https://youtu.be/dQw4w9WgXcQ", true},
		{"https://vimeo.com/123456789", true},
		{"https://tv.naver.com/v/987654321", true},
		{"https://malicious-site.com/video.mp4", false},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			_, err := svc.CreateVideoEmbed(context.Background(), "blog-id", "user-id", VideoEmbedInput{VideoURL: tt.url})
			// Without DB pool, unsupported provider errors before DB query (if validation logic runs)
			if !tt.expected && err == nil {
				t.Fatalf("expected error for URL: %s", tt.url)
			}
		})
	}
}

func TestMediaCursorRoundTrip(t *testing.T) {
	moment := time.Date(2026, 5, 1, 12, 30, 0, 123456789, time.UTC)
	cursor := encodeMediaCursor(moment, "550e8400-e29b-41d4-a716-446655440000")
	backTime, backID, err := decodeMediaCursor(cursor)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !backTime.Equal(moment) || backID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Fatalf("round trip mismatch: %v %s", backTime, backID)
	}
	for _, bad := range []string{"", "!!!", "bm90LWEtdGltZXQ="} {
		if _, _, err := decodeMediaCursor(bad); err == nil {
			t.Fatalf("bad cursor %q accepted", bad)
		}
	}
}

func TestNewServiceRequiresPool(t *testing.T) {
	if _, err := NewService(nil, StorageConfig{}); err == nil {
		t.Fatalf("nil pool accepted")
	}
}
