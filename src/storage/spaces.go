package storage

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
	"vatusa-cobalt/config"
)

// ErrNotConfigured is returned when an upload is attempted on an instance that
// has no Spaces credentials (e.g. a local dev environment).
var ErrNotConfigured = errors.New("object storage is not configured")

var httpClient = &http.Client{Timeout: 30 * time.Second}

// UploadEventBanner validates an uploaded banner and stores it in the Spaces
// bucket, returning the public URL to record on the event. Callers should run
// their permission checks first — this writes to the bucket unconditionally.
func UploadEventBanner(ctx context.Context, facility string, header *multipart.FileHeader) (string, error) {
	if !config.IsSpacesConfigured() {
		return "", ErrNotConfigured
	}
	if header.Size > MaxBannerBytes {
		return "", fmt.Errorf("%w: file is larger than %d MB", ErrInvalidImage, MaxBannerBytes>>20)
	}

	file, err := header.Open()
	if err != nil {
		return "", fmt.Errorf("%w: could not read uploaded file", ErrInvalidImage)
	}
	defer file.Close()

	// LimitReader guards against a Content-Length that understates the body.
	data, err := io.ReadAll(io.LimitReader(file, MaxBannerBytes+1))
	if err != nil {
		return "", fmt.Errorf("%w: could not read uploaded file", ErrInvalidImage)
	}

	info, err := inspectImage(data)
	if err != nil {
		return "", err
	}

	key := eventBannerKey(facility, info.extension)
	if err := putObject(ctx, config.SpacesEndpoint(), config.SpacesRegion(), key, info.contentType, data); err != nil {
		return "", err
	}

	return config.SpacesPublicBaseURL() + "/" + key, nil
}

// eventBannerKey builds an unguessable, collision-free object key. Keys are
// never reused, so an edit that replaces a banner simply writes a new object
// and leaves the old one behind rather than risking a cache-stale overwrite.
func eventBannerKey(facility, extension string) string {
	facility = strings.ToLower(strings.TrimSpace(facility))
	if facility == "" {
		facility = "unknown"
	}
	facility = sanitizeKeySegment(facility)

	var random [16]byte
	if _, err := rand.Read(random[:]); err != nil {
		// crypto/rand does not fail in practice; fall back to a timestamp so a
		// banner upload never hard-fails on entropy.
		return fmt.Sprintf("event-banners/%s/%d.%s", facility, time.Now().UnixNano(), extension)
	}

	return fmt.Sprintf("event-banners/%s/%s.%s", facility, hex.EncodeToString(random[:]), extension)
}

func sanitizeKeySegment(s string) string {
	var out strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-' || c == '_' {
			out.WriteByte(c)
		}
	}
	if out.Len() == 0 {
		return "unknown"
	}
	return out.String()
}

// putObject uploads data to the given bucket endpoint (e.g.
// config.SpacesEndpoint() or config.DocsEndpoint()), signed for region.
func putObject(ctx context.Context, endpoint, region, key, contentType string, data []byte) error {
	url := endpoint + "/" + key

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("building spaces request: %w", err)
	}

	req.ContentLength = int64(len(data))
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("X-Amz-Acl", "public-read")
	// Keys embed random bytes and are never rewritten, so the object at a
	// given URL is immutable and safe to cache indefinitely at the edge.
	req.Header.Set("Cache-Control", "public, max-age=31536000, immutable")

	signRequest(req, config.SpacesKey(), config.SpacesSecret(), region, "s3", hashHex(data), time.Now())

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("uploading to spaces: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("spaces rejected the upload (%s): %s", resp.Status, strings.TrimSpace(string(body)))
	}

	return nil
}
