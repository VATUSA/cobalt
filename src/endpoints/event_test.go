package endpoints

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
)

func TestParseEventRange(t *testing.T) {
	cases := []struct {
		name     string
		query    string
		wantOk   bool
		wantFrom int64
		wantTo   int64
	}{
		{"valid", "?from=1000&to=2000", true, 1000, 2000},
		{"exactly_max_span", "?from=0&to=31622400", true, 0, 31622400},
		{"over_max_span", "?from=0&to=31622401", false, 0, 0},
		{"to_equals_from", "?from=1000&to=1000", false, 0, 0},
		{"to_before_from", "?from=2000&to=1000", false, 0, 0},
		{"missing_from", "?to=2000", false, 0, 0},
		{"missing_to", "?from=1000", false, 0, 0},
		{"non_numeric", "?from=abc&to=2000", false, 0, 0},
		{"milliseconds_too_wide", "?from=1700000000000&to=1800000000000", false, 0, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/event/range"+tc.query, nil)
			c := echo.NewContext(req, httptest.NewRecorder(), echo.New())
			from, to, err := parseEventRange(c)
			if (err == nil) != tc.wantOk {
				t.Fatalf("err = %v, wantOk %v", err, tc.wantOk)
			}
			if tc.wantOk && (from != tc.wantFrom || to != tc.wantTo) {
				t.Fatalf("got (%d, %d), want (%d, %d)", from, to, tc.wantFrom, tc.wantTo)
			}
		})
	}
}
