package endpoints

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
)

func newTestContext(method, target string, pathValues ...echo.PathValue) *echo.Context {
	req := httptest.NewRequest(method, target, nil)
	rec := httptest.NewRecorder()
	c := echo.NewContext(req, rec, echo.New())
	c.SetPathValues(echo.PathValues(pathValues))
	return c
}

func TestParseId32(t *testing.T) {
	cases := []struct {
		name    string
		raw     string
		wantOk  bool
		wantVal int32
	}{
		{"valid", "42", true, 42},
		{"zero", "0", false, 0},
		{"negative", "-1", false, 0},
		{"non_numeric", "abc", false, 0},
		{"overflow_int32", "4294967297", false, 0}, // 2^32 + 1, would truncate to 1 under int32(strconv.Atoi(...))
		{"empty", "", false, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := newTestContext(http.MethodGet, "/", echo.PathValue{Name: "id", Value: tc.raw})
			got, ok := parseId32(c, "id")
			if ok != tc.wantOk {
				t.Fatalf("ok = %v, want %v", ok, tc.wantOk)
			}
			if ok && got != tc.wantVal {
				t.Errorf("value = %d, want %d", got, tc.wantVal)
			}
		})
	}
}

func TestParseId64(t *testing.T) {
	cases := []struct {
		name   string
		raw    string
		wantOk bool
	}{
		{"valid", "42", true},
		{"zero", "0", false},
		{"negative", "-1", false},
		{"non_numeric", "abc", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := newTestContext(http.MethodGet, "/", echo.PathValue{Name: "id", Value: tc.raw})
			_, ok := parseId64(c, "id")
			if ok != tc.wantOk {
				t.Fatalf("ok = %v, want %v", ok, tc.wantOk)
			}
		})
	}
}

func TestRequireText(t *testing.T) {
	if err := requireText("title", "  hello  ", 10); err != nil {
		t.Errorf("unexpected error for valid text: %v", err)
	}
	if err := requireText("title", "   ", 10); err == nil {
		t.Error("expected error for blank text")
	}
	if err := requireText("title", "this is too long", 5); err == nil {
		t.Error("expected error for over-long text")
	}
}
