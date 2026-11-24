package media

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"

	"golang.org/x/image/webp"
)

// TestVariantWEBPTwin pins the G-12 emission rule: the lossless WebP twin
// ships only when it undercuts its JPEG sibling, and every emitted twin
// decodes with the standard library. Flat graphics win; noisy photos lose.
func TestVariantWEBPTwin(t *testing.T) {
	flat := image.NewRGBA(image.Rect(0, 0, 640, 480))
	for y := 0; y < 480; y++ {
		for x := 0; x < 640; x++ {
			c := color.RGBA{255, 255, 255, 255}
			if (x/40+y/40)%2 == 0 {
				c = color.RGBA{30, 60, 200, 255}
			}
			flat.Set(x, y, c)
		}
	}
	encoded := encodeVariant(flat)
	if encoded == nil || len(encoded.jpeg) == 0 || len(encoded.webp) == 0 {
		t.Fatal("expected jpeg + webp twins for flat graphics")
	}
	if len(encoded.webp) >= len(encoded.jpeg) {
		t.Fatalf("expected webp (%d) to undercut jpeg (%d) on flat graphics", len(encoded.webp), len(encoded.jpeg))
	}
	if _, err := webp.Decode(bytes.NewReader(encoded.webp)); err != nil {
		t.Fatalf("webp twin undecodable: %v", err)
	}

	seed := uint32(12345)
	noisy := image.NewRGBA(image.Rect(0, 0, 320, 240))
	for y := 0; y < 240; y++ {
		for x := 0; x < 320; x++ {
			seed = seed*1103515245 + 12345
			noisy.Set(x, y, color.RGBA{uint8(seed >> 16), uint8(seed >> 8), uint8(seed), 255})
		}
	}
	photo := encodeVariant(noisy)
	if photo == nil || len(photo.jpeg) == 0 {
		t.Fatal("expected jpeg for noisy photo")
	}
	if len(photo.webp) > 0 && len(photo.webp) < len(photo.jpeg) {
		t.Fatalf("unexpected webp win on noise: webp=%d jpeg=%d", len(photo.webp), len(photo.jpeg))
	}
}

func testImageBytes(t *testing.T, format string, width, height int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 128, A: 255})
		}
	}
	var buf bytes.Buffer
	var err error
	if format == "png" {
		err = png.Encode(&buf, img)
	} else {
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
	}
	if err != nil {
		t.Fatalf("encode %s: %v", format, err)
	}
	return buf.Bytes()
}

func TestDecodeImageDimensions(t *testing.T) {
	for _, tc := range []struct{ mime, format string }{
		{"image/jpeg", "jpeg"},
		{"image/png", "png"},
	} {
		body := testImageBytes(t, tc.format, 800, 600)
		width, height, img, err := decodeImage(tc.mime, body)
		if err != nil {
			t.Fatalf("%s: %v", tc.mime, err)
		}
		if width != 800 || height != 600 || img == nil {
			t.Fatalf("%s: got %dx%d img=%v", tc.mime, width, height, img != nil)
		}
	}
}

func TestDecodeImageRejectsCorrupt(t *testing.T) {
	if _, _, _, err := decodeImage("image/jpeg", []byte{0xFF, 0xD8, 0xFF, 0x00}); err == nil {
		t.Fatalf("corrupt jpeg accepted")
	}
	if _, _, _, err := decodeImage("image/avif", bytes.Repeat([]byte{0}, 64)); err == nil {
		t.Fatalf("avif unexpectedly decodable")
	}
}

func TestStripEXIFRoundTrip(t *testing.T) {
	for _, mime := range []string{"image/jpeg", "image/png"} {
		format := "jpeg"
		if mime == "image/png" {
			format = "png"
		}
		_, _, img, err := decodeImage(mime, testImageBytes(t, format, 640, 480))
		if err != nil {
			t.Fatalf("decode: %v", err)
		}
		stripped, err := stripEXIF(mime, img)
		if err != nil {
			t.Fatalf("stripEXIF %s: %v", mime, err)
		}
		width, height, _, err := decodeImage(mime, stripped)
		if err != nil {
			t.Fatalf("re-decode stripped %s: %v", mime, err)
		}
		if width != 640 || height != 480 {
			t.Fatalf("stripped dims = %dx%d", width, height)
		}
	}
	if _, err := stripEXIF("image/webp", image.NewRGBA(image.Rect(0, 0, 8, 8))); err == nil {
		t.Fatalf("webp strip unexpectedly supported")
	}
}

func TestBuildVariantsNeverUpscales(t *testing.T) {
	_, _, large, err := decodeImage("image/jpeg", testImageBytes(t, "jpeg", 2000, 1000))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	variants := buildVariants(large)
	names := make(map[string]bool)
	for _, v := range variants {
		names[v.Name] = true
		if len(v.JPEG) == 0 {
			t.Fatalf("empty jpeg for %s", v.Name)
		}
		if _, _, _, err := decodeImage("image/jpeg", v.JPEG); err != nil {
			t.Fatalf("variant %s undecodable: %v", v.Name, err)
		}
	}
	for _, want := range []string{"THUMB_400", "W320", "W640", "W1200", "W1920"} {
		if !names[want] {
			t.Errorf("missing variant %s (got %v)", want, names)
		}
	}

	_, _, small, err := decodeImage("image/jpeg", testImageBytes(t, "jpeg", 200, 100))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	smallVariants := buildVariants(small)
	if len(smallVariants) != 1 || smallVariants[0].Name != "THUMB_400" {
		names := make([]string, 0, len(smallVariants))
		for _, v := range smallVariants {
			names = append(names, v.Name)
		}
		t.Fatalf("small source must yield only THUMB_400, got %v", names)
	}
}
