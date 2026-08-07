package storage

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func pngBytes(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	img.Set(0, 0, color.RGBA{R: 1, A: 255})
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestInspectImageAccepts16x9(t *testing.T) {
	info, err := inspectImage(pngBytes(t, 1920, 1080))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.contentType != "image/png" || info.extension != "png" {
		t.Errorf("got %q/%q, want image/png/png", info.contentType, info.extension)
	}
	if info.width != 1920 || info.height != 1080 {
		t.Errorf("got %dx%d, want 1920x1080", info.width, info.height)
	}
}

func TestInspectImageRejectsWrongAspectRatio(t *testing.T) {
	_, err := inspectImage(pngBytes(t, 1000, 1000))
	if !errors.Is(err, ErrInvalidImage) {
		t.Fatalf("got %v, want ErrInvalidImage", err)
	}
}

// The check that keeps the public bucket from becoming a stored-XSS vector.
func TestInspectImageRejectsNonImage(t *testing.T) {
	for name, payload := range map[string][]byte{
		"html":  []byte("<html><script>alert(1)</script></html>"),
		"svg":   []byte(`<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`),
		"empty": {},
	} {
		if _, err := inspectImage(payload); !errors.Is(err, ErrInvalidImage) {
			t.Errorf("%s: got %v, want ErrInvalidImage", name, err)
		}
	}
}

func TestInspectImageRejectsOversizedFile(t *testing.T) {
	if _, err := inspectImage(make([]byte, MaxBannerBytes+1)); !errors.Is(err, ErrInvalidImage) {
		t.Fatalf("got %v, want ErrInvalidImage", err)
	}
}

func TestWebpDimensions(t *testing.T) {
	// Minimal VP8X (extended format) container declaring a 1920x1080 canvas;
	// canvas fields are 24-bit little-endian and stored as size-1.
	vp8x := make([]byte, 30)
	copy(vp8x[0:4], "RIFF")
	copy(vp8x[8:12], "WEBP")
	copy(vp8x[12:16], "VP8X")
	w, h := 1920-1, 1080-1
	vp8x[24], vp8x[25], vp8x[26] = byte(w), byte(w>>8), byte(w>>16)
	vp8x[27], vp8x[28], vp8x[29] = byte(h), byte(h>>8), byte(h>>16)

	gotW, gotH, ok := webpDimensions(vp8x)
	if !ok || gotW != 1920 || gotH != 1080 {
		t.Fatalf("got %dx%d ok=%v, want 1920x1080 ok=true", gotW, gotH, ok)
	}

	if _, _, ok := webpDimensions(pngBytes(t, 16, 9)); ok {
		t.Error("png should not parse as webp")
	}
}

func TestEventBannerKey(t *testing.T) {
	key := eventBannerKey("ZDV", "png")
	if want := "event-banners/zdv/"; len(key) <= len(want) || key[:len(want)] != want {
		t.Fatalf("key = %q, want prefix %q", key, want)
	}
	if key[len(key)-4:] != ".png" {
		t.Errorf("key = %q, want .png suffix", key)
	}
	if eventBannerKey("ZDV", "png") == key {
		t.Error("keys should not repeat")
	}
	// Path traversal and separators must not survive into the object key.
	if got := eventBannerKey("../../etc", "png"); got[:len("event-banners/etc/")] != "event-banners/etc/" {
		t.Errorf("key = %q, want sanitized facility segment", got)
	}
}
