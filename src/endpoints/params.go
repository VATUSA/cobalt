package endpoints

import (
	"fmt"
	"mime"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v5"
)

// isMultipartRequest reports whether the request's Content-Type is
// multipart/form-data. Uses mime.ParseMediaType rather than a case-sensitive
// strings.HasPrefix, since HTTP header values are case-insensitive and a
// client sending "Multipart/Form-Data" would otherwise silently fall through
// to the JSON binding path and fail to parse the multipart body.
func isMultipartRequest(c *echo.Context) bool {
	contentType := c.Request().Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}
	return mediaType == "multipart/form-data"
}

// parseId32 parses c.Param(label) as a positive int32 path id. On failure it
// writes the 400 itself and returns false, so callers can just check ok. This
// is stricter than strconv.Atoi followed by int32(...), which silently
// truncates an out-of-range id (e.g. 4294967297) down to a small, unrelated
// row.
func parseId32(c *echo.Context, label string) (int32, bool) {
	raw := c.Param(label)
	id, err := strconv.ParseInt(raw, 10, 32)
	if err != nil || id <= 0 {
		_ = GenericError(c, http.StatusBadRequest, fmt.Errorf("invalid %s", label))
		return 0, false
	}
	return int32(id), true
}

// parseId64 is parseId32's counterpart for bigint-keyed tables (solo_cert).
func parseId64(c *echo.Context, label string) (int64, bool) {
	raw := c.Param(label)
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		_ = GenericError(c, http.StatusBadRequest, fmt.Errorf("invalid %s", label))
		return 0, false
	}
	return id, true
}

// requireText validates that value is non-empty after trimming and fits
// within max bytes, so an over-long field is a 400 instead of a 500 leaking
// a MySQL "Data too long for column" error.
func requireText(name, value string, max int) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fmt.Errorf("%s is required", name)
	}
	if len(trimmed) > max {
		return fmt.Errorf("%s must be %d characters or fewer", name, max)
	}
	return nil
}
