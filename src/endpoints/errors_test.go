package endpoints

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-sql-driver/mysql"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func TestClassifyError(t *testing.T) {
	cases := []struct {
		name       string
		status     int
		err        error
		wantStatus int
		wantMsg    string
		wantCode   ErrorCode
		wantLog    bool
	}{
		{
			name:       "no rows upgrades to not found",
			status:     http.StatusInternalServerError,
			err:        sql.ErrNoRows,
			wantStatus: http.StatusNotFound,
			wantMsg:    "not found",
			wantCode:   CodeNotFound,
		},
		{
			name:   "duplicate entry upgrades to conflict",
			status: http.StatusInternalServerError,
			err: &mysql.MySQLError{
				Number:  1062,
				Message: "Duplicate entry '123' for key 'PRIMARY'",
			},
			wantStatus: http.StatusConflict,
			wantMsg:    "already exists",
			wantCode:   CodeConflict,
		},
		{
			name:   "wrapped duplicate entry is still recognized",
			status: http.StatusInternalServerError,
			err: errors.Join(
				errors.New("failed to add role"),
				&mysql.MySQLError{Number: 1062, Message: "Duplicate entry '5' for key 'PRIMARY'"},
			),
			wantStatus: http.StatusConflict,
			wantMsg:    "already exists",
			wantCode:   CodeConflict,
		},
		{
			name:   "foreign key violation is a conflict",
			status: http.StatusInternalServerError,
			err: &mysql.MySQLError{
				Number:  1452,
				Message: "Cannot add or update a child row: a foreign key constraint fails",
			},
			wantStatus: http.StatusConflict,
			wantMsg:    "conflict",
			wantCode:   CodeConflict,
		},
		{
			name:   "deadlock is a retryable conflict and is logged",
			status: http.StatusInternalServerError,
			err: &mysql.MySQLError{
				Number:  1213,
				Message: "Deadlock found when trying to get lock; try restarting transaction",
			},
			wantStatus: http.StatusConflict,
			wantMsg:    "conflict",
			wantCode:   CodeConflict,
			wantLog:    true,
		},
		{
			name:   "unknown mysql error is scrubbed and logged",
			status: http.StatusInternalServerError,
			err: &mysql.MySQLError{
				Number:  1045,
				Message: "Access denied for user 'cobalt'@'host' (using password: YES)",
			},
			wantStatus: http.StatusInternalServerError,
			wantMsg:    internalErrorMessage,
			wantCode:   CodeInternal,
			wantLog:    true,
		},
		{
			name:       "connection error is scrubbed and logged",
			status:     http.StatusInternalServerError,
			err:        &net.OpError{Op: "dial", Net: "tcp", Err: errors.New("connection refused")},
			wantStatus: http.StatusInternalServerError,
			wantMsg:    internalErrorMessage,
			wantCode:   CodeInternal,
			wantLog:    true,
		},
		{
			name:       "stale connection is scrubbed and logged",
			status:     http.StatusInternalServerError,
			err:        driver.ErrBadConn,
			wantStatus: http.StatusInternalServerError,
			wantMsg:    internalErrorMessage,
			wantCode:   CodeInternal,
			wantLog:    true,
		},
		{
			name:       "query deadline exceeded is scrubbed and logged",
			status:     http.StatusInternalServerError,
			err:        context.DeadlineExceeded,
			wantStatus: http.StatusInternalServerError,
			wantMsg:    internalErrorMessage,
			wantCode:   CodeInternal,
			wantLog:    true,
		},
		{
			name:       "safe client error passes through verbatim",
			status:     http.StatusBadRequest,
			err:        errors.New("q is required"),
			wantStatus: http.StatusBadRequest,
			wantMsg:    "q is required",
			wantCode:   CodeBadRequest,
		},
		{
			name:       "forbidden status maps to forbidden code",
			status:     http.StatusForbidden,
			err:        errors.New("permission denied"),
			wantStatus: http.StatusForbidden,
			wantMsg:    "permission denied",
			wantCode:   CodeForbidden,
		},
		{
			name:       "not found status maps to not found code",
			status:     http.StatusNotFound,
			err:        errors.New("user not found"),
			wantStatus: http.StatusNotFound,
			wantMsg:    "user not found",
			wantCode:   CodeNotFound,
		},
		{
			name:       "safe message survives a 500 with internal code",
			status:     http.StatusInternalServerError,
			err:        SafeError("failed to update post"),
			wantStatus: http.StatusInternalServerError,
			wantMsg:    "failed to update post",
			wantCode:   CodeInternal,
		},
		{
			name:       "safe message passes through a 4xx verbatim",
			status:     http.StatusBadRequest,
			err:        SafeError("q is required"),
			wantStatus: http.StatusBadRequest,
			wantMsg:    "q is required",
			wantCode:   CodeBadRequest,
		},
		{
			name:       "wrapped safe message is still recognized",
			status:     http.StatusInternalServerError,
			err:        errors.Join(errors.New("context"), SafeError("failed to create token")),
			wantStatus: http.StatusInternalServerError,
			wantMsg:    "failed to create token",
			wantCode:   CodeInternal,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := classifyError(tc.status, tc.err)
			if got.status != tc.wantStatus {
				t.Errorf("status = %d, want %d", got.status, tc.wantStatus)
			}
			if got.message != tc.wantMsg {
				t.Errorf("message = %q, want %q", got.message, tc.wantMsg)
			}
			if got.code != tc.wantCode {
				t.Errorf("code = %q, want %q", got.code, tc.wantCode)
			}
			if got.log != tc.wantLog {
				t.Errorf("log = %v, want %v", got.log, tc.wantLog)
			}
		})
	}
}

func TestRespondErrorScrubsDatabaseErrors(t *testing.T) {
	e := echo.New()
	e.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	e.GET("/scrub", func(c *echo.Context) error {
		return RespondError(c, http.StatusInternalServerError, &mysql.MySQLError{
			Number:  1045,
			Message: "Access denied for user 'cobalt'@'host' (using password: YES)",
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/scrub", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	body := rec.Body.String()
	if strings.Contains(body, "Access denied") {
		t.Errorf("response leaked driver error text: %s", body)
	}
	if !strings.Contains(body, internalErrorMessage) {
		t.Errorf("response missing generic message: %s", body)
	}
	if !strings.Contains(body, `"code":"internal"`) {
		t.Errorf("response missing code: %s", body)
	}
}

func TestRespondErrorUpgradesNotFound(t *testing.T) {
	e := echo.New()
	e.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	e.GET("/upgrade", func(c *echo.Context) error {
		return RespondError(c, http.StatusInternalServerError, sql.ErrNoRows)
	})

	req := httptest.NewRequest(http.MethodGet, "/upgrade", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "not found") || !strings.Contains(body, `"code":"not_found"`) {
		t.Errorf("response = %s", body)
	}
}

func TestRespondErrorPassesThroughClientErrors(t *testing.T) {
	e := echo.New()
	e.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	e.GET("/pass", func(c *echo.Context) error {
		return RespondError(c, http.StatusBadRequest, errors.New("q is required"))
	})

	req := httptest.NewRequest(http.MethodGet, "/pass", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "q is required") {
		t.Errorf("client error message was not preserved: %s", body)
	}
	if !strings.Contains(body, `"code":"bad_request"`) {
		t.Errorf("response missing code: %s", body)
	}
}

func TestRespondErrorPreservesIntentionalMessages(t *testing.T) {
	e := echo.New()
	e.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	e.GET("/safe", func(c *echo.Context) error {
		return RespondError(c, http.StatusInternalServerError, SafeError("failed to update post"))
	})

	req := httptest.NewRequest(http.MethodGet, "/safe", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "failed to update post") {
		t.Errorf("safe message was not preserved: %s", body)
	}
	if strings.Contains(body, internalErrorMessage) {
		t.Errorf("safe message was scrubbed: %s", body)
	}
	if !strings.Contains(body, `"code":"internal"`) {
		t.Errorf("response missing code: %s", body)
	}
}

func TestRespondErrorLogsDiagnosticsWithoutSerializing(t *testing.T) {
	e := echo.New()
	e.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	e.GET("/diag", func(c *echo.Context) error {
		return RespondError(c, http.StatusInternalServerError,
			SafeError("error storing vatsim user record"),
			&mysql.MySQLError{Number: 1045, Message: "Access denied for user 'cobalt'@'host' (using password: YES)"})
	})

	req := httptest.NewRequest(http.MethodGet, "/diag", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "error storing vatsim user record") {
		t.Errorf("first error was not serialized: %s", body)
	}
	if strings.Contains(body, "Access denied") {
		t.Errorf("diagnostic error leaked into the payload: %s", body)
	}
	if !strings.Contains(body, `"code":"internal"`) {
		t.Errorf("response missing code: %s", body)
	}
}

func TestErrorHandlerScrubsPanics(t *testing.T) {
	e := echo.New()
	e.HTTPErrorHandler = ErrorHandler
	e.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	e.Use(middleware.Recover())
	e.GET("/panic", func(c *echo.Context) error {
		panic("boom")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	body := rec.Body.String()
	if strings.Contains(body, "boom") {
		t.Errorf("panic detail leaked: %s", body)
	}
	if !strings.Contains(body, internalErrorMessage) || !strings.Contains(body, `"code":"internal"`) {
		t.Errorf("response = %s", body)
	}
}
