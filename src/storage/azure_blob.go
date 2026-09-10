package storage

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// putObjectAzure uploads data to an Azure Blob Storage container using the
// Shared Key authorization scheme — the Blob-service equivalent of
// putObjectSpaces' SigV4 in sigv4.go. Hand-rolled for the same reason noted
// on sigv4.go: this package makes exactly one kind of call (PUT a whole
// block blob in one shot), so pulling in the full Azure SDK is a lot of
// dependency weight for one HTTP request, and Shared Key's string-to-sign is
// considerably simpler than SigV4's to begin with.
//
// https://learn.microsoft.com/en-us/rest/api/storageservices/authorize-with-shared-key
func putObjectAzure(ctx context.Context, endpoint, account, accountKey, key, contentType string, data []byte) error {
	release, err := acquireUploadSlot(ctx)
	if err != nil {
		return fmt.Errorf("waiting for an upload slot: %w", err)
	}
	defer release()

	url := endpoint + "/" + key

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("building azure blob request: %w", err)
	}

	req.ContentLength = int64(len(data))
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("x-ms-blob-type", "BlockBlob")
	req.Header.Set("x-ms-date", time.Now().UTC().Format(http.TimeFormat))
	req.Header.Set("x-ms-version", "2021-08-06")
	// Same rationale as putObjectSpaces: arbitrary user-uploaded file types
	// and never-rewritten, unguessable keys.
	req.Header.Set("X-Content-Type-Options", "nosniff")
	req.Header.Set("Cache-Control", "public, max-age=31536000, immutable")

	if err := signAzureBlobRequest(req, account, accountKey, key, len(data)); err != nil {
		return fmt.Errorf("signing azure blob request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("uploading to azure blob: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("azure blob rejected the upload (%s): %s", resp.Status, strings.TrimSpace(string(body)))
	}

	return nil
}

// signAzureBlobRequest implements Shared Key authorization for a single PUT
// Blob request and sets the resulting Authorization header on req.
// blobPath is "<container>/<key>" (no leading slash).
func signAzureBlobRequest(req *http.Request, account, accountKey, blobPath string, contentLength int) error {
	keyBytes, err := base64.StdEncoding.DecodeString(accountKey)
	if err != nil {
		return fmt.Errorf("decoding azure storage account key: %w", err)
	}

	// CanonicalizedHeaders: every x-ms-* header, lowercase name, sorted,
	// "name:value\n" each — blob-type < date < version alphabetically, so
	// this fixed order matches what sorting them would produce.
	canonicalizedHeaders := fmt.Sprintf(
		"x-ms-blob-type:%s\nx-ms-date:%s\nx-ms-version:%s\n",
		req.Header.Get("x-ms-blob-type"),
		req.Header.Get("x-ms-date"),
		req.Header.Get("x-ms-version"),
	)
	// CanonicalizedResource: account + the resource path, no query string
	// (this request has none).
	canonicalizedResource := fmt.Sprintf("/%s/%s", account, blobPath)

	contentLengthField := ""
	if contentLength > 0 {
		contentLengthField = strconv.Itoa(contentLength)
	}

	stringToSign := strings.Join([]string{
		http.MethodPut,                 // Verb
		"",                             // Content-Encoding
		"",                             // Content-Language
		contentLengthField,             // Content-Length
		"",                             // Content-MD5
		req.Header.Get("Content-Type"), // Content-Type
		"",                             // Date (x-ms-date is used instead)
		"",                             // If-Modified-Since
		"",                             // If-Match
		"",                             // If-None-Match
		"",                             // If-Unmodified-Since
		"",                             // Range
	}, "\n") + "\n" + canonicalizedHeaders + canonicalizedResource

	mac := hmac.New(sha256.New, keyBytes)
	mac.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	req.Header.Set("Authorization", fmt.Sprintf("SharedKey %s:%s", account, signature))
	return nil
}
