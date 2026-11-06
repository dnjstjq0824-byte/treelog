package media

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"
)

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
