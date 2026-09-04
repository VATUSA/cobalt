package storage

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"vatusa-cobalt/config"
)

// ErrInvalidDocument marks a rejection caused by what the user uploaded
// rather than by a storage failure, so callers can answer 400 instead of 500.
var ErrInvalidDocument = errors.New("invalid document")

// MaxDocumentBytes caps an uploaded policy document. Bigger than a banner —
// the legacy bucket already holds a multi-MB logo asset zip.
const MaxDocumentBytes = 50 << 20

// documentContentTypes is the fixed extension allowlist for policy document
// uploads. Content-Type served is always derived from this map, never from
// the client-supplied filename or Content-Type header.
var documentContentTypes = map[string]string{
	"pdf":  "application/pdf",
	"doc":  "application/msword",
	"docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	"xls":  "application/vnd.ms-excel",
	"xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	"ppt":  "application/vnd.ms-powerpoint",
	"pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
	"zip":  "application/zip",
	"jpg":  "image/jpeg",
	"jpeg": "image/jpeg",
	"png":  "image/png",
	"gif":  "image/gif",
	"txt":  "text/plain; charset=utf-8",
	"csv":  "text/csv; charset=utf-8",
	"md":   "text/markdown; charset=utf-8",
}

// UploadPolicyDocument validates an uploaded policy document and stores it in
// the docs Spaces bucket, returning the public URL to record on the document.
// Callers should run their permission checks first — this writes to the
// bucket unconditionally.
func UploadPolicyDocument(ctx context.Context, header *multipart.FileHeader) (string, error) {
	if !config.IsDocsConfigured() {
		return "", ErrNotConfigured
	}
	if header.Size > MaxDocumentBytes {
		return "", fmt.Errorf("%w: file is larger than %d MB", ErrInvalidDocument, MaxDocumentBytes>>20)
	}

	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(header.Filename), "."))
	contentType, ok := documentContentTypes[ext]
	if !ok {
		return "", fmt.Errorf("%w: unsupported file type %q", ErrInvalidDocument, ext)
	}

	file, err := header.Open()
	if err != nil {
		return "", fmt.Errorf("%w: could not read uploaded file", ErrInvalidDocument)
	}
	defer file.Close()

	// LimitReader guards against a Content-Length that understates the body.
	data, err := io.ReadAll(io.LimitReader(file, MaxDocumentBytes+1))
	if err != nil {
		return "", fmt.Errorf("%w: could not read uploaded file", ErrInvalidDocument)
	}
	if len(data) == 0 {
		return "", fmt.Errorf("%w: file is empty", ErrInvalidDocument)
	}
	if len(data) > MaxDocumentBytes {
		return "", fmt.Errorf("%w: file is larger than %d MB", ErrInvalidDocument, MaxDocumentBytes>>20)
	}

	if err := inspectDocument(ext, data); err != nil {
		return "", err
	}

	key := policyDocumentKey(ext)
	if err := putObject(ctx, config.DocsEndpoint(), config.DocsRegion(), key, contentType, data); err != nil {
		return "", err
	}

	return config.DocsPublicBaseURL() + "/" + key, nil
}

// inspectDocument does lightweight magic-byte validation for the formats that
// have a reliable one. This is the security-relevant step for the same
// reason as inspectImage: without it, an attacker could upload HTML/SVG
// disguised as e.g. a .pdf to a public bucket and get back a trusted-looking
// URL, which is a stored-XSS primitive. txt/csv/md have no reliable magic
// number and are accepted as-is — the lowest-risk formats in the allowlist.
func inspectDocument(ext string, data []byte) error {
	switch ext {
	case "pdf":
		if !bytes.HasPrefix(data, []byte("%PDF-")) {
			return fmt.Errorf("%w: file is not a valid PDF", ErrInvalidDocument)
		}
	case "zip", "docx", "xlsx", "pptx":
		// docx/xlsx/pptx are all zip containers under the hood.
		if !bytes.HasPrefix(data, []byte("PK")) {
			return fmt.Errorf("%w: file is not a valid %s", ErrInvalidDocument, ext)
		}
	case "doc", "xls", "ppt":
		oleMagic := []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}
		if !bytes.HasPrefix(data, oleMagic) {
			return fmt.Errorf("%w: file is not a valid %s", ErrInvalidDocument, ext)
		}
	case "jpg", "jpeg", "png", "gif":
		cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
		if err != nil || cfg.Width <= 0 || cfg.Height <= 0 {
			return fmt.Errorf("%w: file is not a valid %s image", ErrInvalidDocument, ext)
		}
		wantFormat := ext
		if wantFormat == "jpg" {
			wantFormat = "jpeg"
		}
		if format != wantFormat {
			return fmt.Errorf("%w: file content does not match its %s extension", ErrInvalidDocument, ext)
		}
	}
	return nil
}

// policyDocumentKey builds an unguessable, collision-free object key under
// the docs/ prefix already live in the vatusa-storage bucket. Keys are never
// reused, so an edit that replaces a document simply writes a new object and
// leaves the old one behind rather than risking a cache-stale overwrite.
//
// rand.Read panics rather than returning an error as of Go 1.24 (it reads
// from the OS CSPRNG, which cannot meaningfully fail at runtime), so there is
// no fallback path to write here — one that swapped the unguessable key for
// a predictable timestamp would defeat the property this function exists for.
func policyDocumentKey(extension string) string {
	var random [16]byte
	_, _ = rand.Read(random[:])
	return fmt.Sprintf("docs/%s.%s", hex.EncodeToString(random[:]), extension)
}
