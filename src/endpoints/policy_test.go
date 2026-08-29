package endpoints

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"vatusa-cobalt/models"

	"github.com/labstack/echo/v5"
)

func multipartPolicyRequest(t *testing.T, fields map[string]string) *echo.Context {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for name, value := range fields {
		if err := writer.WriteField(name, value); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/policy/document", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	return echo.NewContext(req, rec, echo.New())
}

func baseFields() map[string]string {
	return map[string]string{
		"policy_category_id": "1",
		"ident":              "test-ident",
		"title":              "Test Title",
		"summary":            "a summary",
		"document_url":       "https://example.test/doc.pdf",
		"effective_date":     "2026-01-01",
	}
}

// TestBindPolicyDocumentRequestHiddenAbsentVsFalse covers the fix for a bug
// where an omitted "hidden" field on a multipart update was indistinguishable
// from an explicit false, which caused an update that didn't touch
// visibility to silently un-hide (or, with the original bug, re-publish) a
// hidden policy document. With Hidden as *bool, absent must decode to nil so
// the caller can fall back to the existing value.
func TestBindPolicyDocumentRequestHiddenAbsentVsFalse(t *testing.T) {
	t.Run("absent", func(t *testing.T) {
		c := multipartPolicyRequest(t, baseFields())
		var request models.PolicyDocumentRequest
		if _, err := bindPolicyDocumentRequest(c, &request); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if request.Hidden != nil {
			t.Errorf("Hidden = %v, want nil when the field is absent", *request.Hidden)
		}
	})

	t.Run("explicit_false", func(t *testing.T) {
		fields := baseFields()
		fields["hidden"] = "false"
		c := multipartPolicyRequest(t, fields)
		var request models.PolicyDocumentRequest
		if _, err := bindPolicyDocumentRequest(c, &request); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if request.Hidden == nil || *request.Hidden != false {
			t.Errorf("Hidden = %v, want explicit false", request.Hidden)
		}
	})

	t.Run("explicit_true", func(t *testing.T) {
		fields := baseFields()
		fields["hidden"] = "true"
		c := multipartPolicyRequest(t, fields)
		var request models.PolicyDocumentRequest
		if _, err := bindPolicyDocumentRequest(c, &request); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if request.Hidden == nil || *request.Hidden != true {
			t.Errorf("Hidden = %v, want explicit true", request.Hidden)
		}
	})

	t.Run("malformed", func(t *testing.T) {
		fields := baseFields()
		fields["hidden"] = "not-a-bool"
		c := multipartPolicyRequest(t, fields)
		var request models.PolicyDocumentRequest
		if _, err := bindPolicyDocumentRequest(c, &request); err == nil {
			t.Error("expected an error for a malformed hidden value, got nil")
		}
	})
}

func TestBindPolicyDocumentRequestSortOrderAbsentVsZero(t *testing.T) {
	t.Run("absent", func(t *testing.T) {
		c := multipartPolicyRequest(t, baseFields())
		var request models.PolicyDocumentRequest
		if _, err := bindPolicyDocumentRequest(c, &request); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if request.SortOrder != nil {
			t.Errorf("SortOrder = %v, want nil when the field is absent", *request.SortOrder)
		}
	})

	t.Run("explicit_zero", func(t *testing.T) {
		fields := baseFields()
		fields["sort_order"] = "0"
		c := multipartPolicyRequest(t, fields)
		var request models.PolicyDocumentRequest
		if _, err := bindPolicyDocumentRequest(c, &request); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if request.SortOrder == nil || *request.SortOrder != 0 {
			t.Errorf("SortOrder = %v, want explicit 0", request.SortOrder)
		}
	})

	t.Run("malformed", func(t *testing.T) {
		fields := baseFields()
		fields["sort_order"] = "not-a-number"
		c := multipartPolicyRequest(t, fields)
		var request models.PolicyDocumentRequest
		if _, err := bindPolicyDocumentRequest(c, &request); err == nil {
			t.Error("expected an error for a malformed sort_order value, got nil")
		}
	})
}

func TestBindPolicyDocumentRequestMalformedCategoryId(t *testing.T) {
	fields := baseFields()
	fields["policy_category_id"] = "not-a-number"
	c := multipartPolicyRequest(t, fields)
	var request models.PolicyDocumentRequest
	if _, err := bindPolicyDocumentRequest(c, &request); err == nil {
		t.Error("expected an error for a malformed policy_category_id, got nil")
	}
}

func TestBindPolicyDocumentRequestZeroByteFileRejected(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for name, value := range baseFields() {
		if err := writer.WriteField(name, value); err != nil {
			t.Fatal(err)
		}
	}
	part, err := writer.CreateFormFile(documentFormField, "empty.pdf")
	if err != nil {
		t.Fatal(err)
	}
	_ = part // zero bytes written
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/policy/document", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	c := echo.NewContext(req, rec, echo.New())

	var request models.PolicyDocumentRequest
	if _, err := bindPolicyDocumentRequest(c, &request); err == nil {
		t.Error("expected an error for a zero-byte uploaded file, got nil")
	}
}
