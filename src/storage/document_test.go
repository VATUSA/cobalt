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

// documentUpload builds the *multipart.FileHeader an endpoint would hand us,
// by round-tripping through a real multipart body.
func documentUpload(t *testing.T, filename string, content []byte) *multipart.FileHeader {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("document", filename)
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
	if err := req.ParseMultipartForm(MaxDocumentBytes); err != nil {
		t.Fatal(err)
	}

	return req.MultipartForm.File["document"][0]
}

func configureDocs(t *testing.T, endpoint, publicBase string) {
	t.Helper()
	t.Setenv("DO_SPACES_KEY", "TESTACCESSKEY")
	t.Setenv("DO_SPACES_SECRET", "testsecret")
	t.Setenv("DO_SPACES_DOCS_REGION", "nyc3")
	t.Setenv("DO_SPACES_DOCS_BUCKET", "vatusa-storage")
	t.Setenv("DO_SPACES_DOCS_ENDPOINT", endpoint)
	t.Setenv("DO_SPACES_DOCS_PUBLIC_BASE_URL", publicBase)
}

func TestUploadPolicyDocumentRoundTrip(t *testing.T) {
	pdf := append([]byte("%PDF-1.4\n"), []byte("rest of a fake but valid-enough pdf")...)

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

	configureDocs(t, server.URL, "https://cdn.example.test")

	url, err := UploadPolicyDocument(context.Background(), documentUpload(t, "policy.pdf", pdf))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(url, "https://cdn.example.test/docs/") {
		t.Errorf("url = %q, want CDN base and docs/ key", url)
	}
	if !strings.HasSuffix(url, ".pdf") {
		t.Errorf("url = %q, want .pdf suffix", url)
	}
	if want := strings.TrimPrefix(url, "https://cdn.example.test"); gotPath != want {
		t.Errorf("uploaded to %q, but returned a URL for %q", gotPath, want)
	}
	if !bytes.Equal(gotBody, pdf) {
		t.Errorf("uploaded %d bytes, want the original %d", len(gotBody), len(pdf))
	}
	if gotACL != "public-read" {
		t.Errorf("x-amz-acl = %q, want public-read (documents are served publicly)", gotACL)
	}
	if gotContentType != "application/pdf" {
		t.Errorf("content-type = %q, want application/pdf", gotContentType)
	}
	if !strings.Contains(gotCacheControl, "immutable") {
		t.Errorf("cache-control = %q, want an immutable directive", gotCacheControl)
	}
	if !strings.Contains(gotAuth, "Credential=TESTACCESSKEY/") || !strings.Contains(gotAuth, "/nyc3/s3/aws4_request") {
		t.Errorf("authorization = %q, want a SigV4 header scoped to nyc3/s3", gotAuth)
	}
}

func TestUploadPolicyDocumentAcceptsZipBasedFormats(t *testing.T) {
	docx := append([]byte("PK\x03\x04"), []byte("rest of a fake but valid-enough docx")...)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	configureDocs(t, server.URL, server.URL)

	url, err := UploadPolicyDocument(context.Background(), documentUpload(t, "handbook.docx", docx))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasSuffix(url, ".docx") {
		t.Errorf("url = %q, want .docx suffix", url)
	}
}

func TestUploadPolicyDocumentRejectsSpoofedExtensionWithoutCallingSpaces(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	configureDocs(t, server.URL, server.URL)

	_, err := UploadPolicyDocument(context.Background(), documentUpload(t, "evil.pdf", []byte("<script>alert(1)</script>")))
	if !errors.Is(err, ErrInvalidDocument) {
		t.Fatalf("got %v, want ErrInvalidDocument", err)
	}
	if called {
		t.Error("a rejected upload must not reach the bucket")
	}
}

func TestUploadPolicyDocumentRejectsUnknownExtension(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	configureDocs(t, server.URL, server.URL)

	_, err := UploadPolicyDocument(context.Background(), documentUpload(t, "script.exe", []byte("MZ")))
	if !errors.Is(err, ErrInvalidDocument) {
		t.Fatalf("got %v, want ErrInvalidDocument", err)
	}
}

func TestUploadPolicyDocumentRequiresConfiguration(t *testing.T) {
	t.Setenv("DO_SPACES_KEY", "")
	t.Setenv("DO_SPACES_SECRET", "")
	t.Setenv("DO_SPACES_DOCS_BUCKET", "")

	pdf := append([]byte("%PDF-1.4\n"), []byte("rest")...)
	_, err := UploadPolicyDocument(context.Background(), documentUpload(t, "policy.pdf", pdf))
	if !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("got %v, want ErrNotConfigured", err)
	}
}
