package storage

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// bannerUpload builds the *multipart.FileHeader an endpoint would hand us, by
// round-tripping through a real multipart body.
func bannerUpload(t *testing.T, filename string, content []byte) *multipart.FileHeader {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("banner_image", filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if err := req.ParseMultipartForm(MaxBannerBytes); err != nil {
		t.Fatal(err)
	}

	return req.MultipartForm.File["banner_image"][0]
}

func configureSpaces(t *testing.T, endpoint, publicBase string) {
	t.Helper()
	t.Setenv("DO_SPACES_KEY", "TESTACCESSKEY")
	t.Setenv("DO_SPACES_SECRET", "testsecret")
	t.Setenv("DO_SPACES_REGION", "nyc3")
	t.Setenv("DO_SPACES_BUCKET", "vatusa-events")
	t.Setenv("DO_SPACES_ENDPOINT", endpoint)
	t.Setenv("DO_SPACES_PUBLIC_BASE_URL", publicBase)
}

func TestUploadEventBannerRoundTrip(t *testing.T) {
	image := pngBytes(t, 1920, 1080)

	var gotPath, gotACL, gotContentType, gotCacheControl, gotAuth string
	var gotBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotACL = r.Header.Get("X-Amz-Acl")
		gotContentType = r.Header.Get("Content-Type")
		gotCacheControl = r.Header.Get("Cache-Control")
		gotAuth = r.Header.Get("Authorization")
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	configureSpaces(t, server.URL, "https://cdn.example.test")

	url, err := UploadEventBanner(context.Background(), "ZDV", bannerUpload(t, "banner.png", image))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(url, "https://cdn.example.test/event-banners/zdv/") {
		t.Errorf("url = %q, want CDN base and facility-scoped key", url)
	}
	if !strings.HasSuffix(url, ".png") {
		t.Errorf("url = %q, want .png suffix", url)
	}
	// The returned URL must address exactly the object we wrote.
	if want := strings.TrimPrefix(url, "https://cdn.example.test"); gotPath != want {
		t.Errorf("uploaded to %q, but returned a URL for %q", gotPath, want)
	}
	if !bytes.Equal(gotBody, image) {
		t.Errorf("uploaded %d bytes, want the original %d", len(gotBody), len(image))
	}
	if gotACL != "public-read" {
		t.Errorf("x-amz-acl = %q, want public-read (banners are served publicly)", gotACL)
	}
	if gotContentType != "image/png" {
		t.Errorf("content-type = %q, want image/png", gotContentType)
	}
	if !strings.Contains(gotCacheControl, "immutable") {
		t.Errorf("cache-control = %q, want an immutable directive", gotCacheControl)
	}
	if !strings.Contains(gotAuth, "Credential=TESTACCESSKEY/") || !strings.Contains(gotAuth, "/nyc3/s3/aws4_request") {
		t.Errorf("authorization = %q, want a SigV4 header scoped to nyc3/s3", gotAuth)
	}
}

// The content type stored is derived from the bytes, not from the filename or
// the browser-supplied part header.
func TestUploadEventBannerIgnoresClaimedFilename(t *testing.T) {
	var gotPath, gotContentType string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotContentType = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	configureSpaces(t, server.URL, server.URL)

	if _, err := UploadEventBanner(context.Background(), "ZDV", bannerUpload(t, "evil.html", pngBytes(t, 1600, 900))); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasSuffix(gotPath, ".png") || gotContentType != "image/png" {
		t.Errorf("stored %q as %q, want a .png key and image/png", gotPath, gotContentType)
	}
}

func TestUploadEventBannerRejectsNonImageWithoutCallingSpaces(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	configureSpaces(t, server.URL, server.URL)

	_, err := UploadEventBanner(context.Background(), "ZDV", bannerUpload(t, "x.png", []byte("<script>alert(1)</script>")))
	if !errors.Is(err, ErrInvalidImage) {
		t.Fatalf("got %v, want ErrInvalidImage", err)
	}
	if called {
		t.Error("a rejected upload must not reach the bucket")
	}
}

func TestUploadEventBannerSurfacesSpacesFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("<Error><Code>SignatureDoesNotMatch</Code></Error>"))
	}))
	defer server.Close()

	configureSpaces(t, server.URL, server.URL)

	_, err := UploadEventBanner(context.Background(), "ZDV", bannerUpload(t, "b.png", pngBytes(t, 1920, 1080)))
	if err == nil {
		t.Fatal("expected an error")
	}
	if errors.Is(err, ErrInvalidImage) {
		t.Error("a bucket failure must not be reported as a bad upload (it would answer 400 instead of 500)")
	}
	if !strings.Contains(err.Error(), "SignatureDoesNotMatch") {
		t.Errorf("error = %v, want the bucket's response included for debugging", err)
	}
}

func TestUploadEventBannerRequiresConfiguration(t *testing.T) {
	t.Setenv("DO_SPACES_KEY", "")
	t.Setenv("DO_SPACES_SECRET", "")
	t.Setenv("DO_SPACES_BUCKET", "")

	_, err := UploadEventBanner(context.Background(), "ZDV", bannerUpload(t, "b.png", pngBytes(t, 1920, 1080)))
	if !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("got %v, want ErrNotConfigured", err)
	}
}
