package endpoints

import (
	"database/sql"
	"errors"
	"net/http"
	"vatusa-cobalt/auth"
	"vatusa-cobalt/models"

	"github.com/go-sql-driver/mysql"
	"github.com/labstack/echo/v5"
)

// ErrorCode lets clients branch on a response without string-matching prose.
type ErrorCode string

const (
	CodeInternal            ErrorCode = "internal"
	CodeNotFound            ErrorCode = "not_found"
	CodeConflict            ErrorCode = "conflict"
	CodeBadRequest          ErrorCode = "bad_request"
	CodeForbidden           ErrorCode = "forbidden"
	CodeUnauthorized        ErrorCode = "unauthorized"
	CodeMethodNotAllowed    ErrorCode = "method_not_allowed"
	CodeUnprocessableEntity ErrorCode = "unprocessable_entity"
)

const internalErrorMessage = "internal server error"

type safeError struct {
	message string
}

func (e *safeError) Error() string {
	return e.message
}

// SafeError marks a message as intentionally client-visible, so it is returned
// verbatim instead of being scrubbed.
func SafeError(message string) error {
	return &safeError{message: message}
}

type classifiedError struct {
	status  int
	message string
	code    ErrorCode
	log     bool
}

func codeForStatus(status int) ErrorCode {
	switch status {
	case http.StatusNotFound:
		return CodeNotFound
	case http.StatusConflict:
		return CodeConflict
	case http.StatusForbidden:
		return CodeForbidden
	case http.StatusUnauthorized:
		return CodeUnauthorized
	case http.StatusMethodNotAllowed:
		return CodeMethodNotAllowed
	case http.StatusUnprocessableEntity:
		return CodeUnprocessableEntity
	default:
		return CodeBadRequest
	}
}

// classifyError maps an error to what the client may see. SafeError and 4xx
// messages pass through verbatim; known database conditions map to a real
// status; everything else is a scrubbed, logged 500.
func classifyError(status int, err error) classifiedError {
	var safe *safeError
	if errors.As(err, &safe) {
		code := CodeInternal
		if status < http.StatusInternalServerError {
			code = codeForStatus(status)
		}
		// A safe message at 5xx is still a server failure: log it so a
		// dropped underlying error can't make the failure invisible.
		return classifiedError{status: status, message: safe.message, code: code, log: status >= http.StatusInternalServerError}
	}

	if status < http.StatusInternalServerError {
		return classifiedError{status: status, message: err.Error(), code: codeForStatus(status)}
	}

	var mysqlErr *mysql.MySQLError
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return classifiedError{status: http.StatusNotFound, message: "not found", code: CodeNotFound}
	case errors.As(err, &mysqlErr):
		switch mysqlErr.Number {
		case 1062:
			return classifiedError{status: http.StatusConflict, message: "already exists", code: CodeConflict, log: true}
		case 1213:
			return classifiedError{status: http.StatusConflict, message: "conflict", code: CodeConflict, log: true}
		case 1451, 1452:
			return classifiedError{status: http.StatusConflict, message: "conflict", code: CodeConflict, log: true}
		default:
			return classifiedError{status: http.StatusInternalServerError, message: internalErrorMessage, code: CodeInternal, log: true}
		}
	default:
		return classifiedError{status: http.StatusInternalServerError, message: internalErrorMessage, code: CodeInternal, log: true}
	}
}

// logScrubbedError records the full failure with request context, since the
// scrubbed response no longer carries it.
func logScrubbedError(c *echo.Context, err error) {
	c.Logger().Error("request failed",
		"method", c.Request().Method,
		"path", c.Request().URL.Path,
		"cid", auth.GetUserCid(c),
		"error", err,
	)
}

// ErrorHandler renders errors that escape handlers (panics, or a handler
// returning an error without writing a response), using the same classification
// as RespondError.
func ErrorHandler(c *echo.Context, err error) {
	status := http.StatusInternalServerError
	var httpErr *echo.HTTPError
	if errors.As(err, &httpErr) && httpErr.Code != 0 {
		status = httpErr.Code
	}
	classified := classifyError(status, err)
	if classified.log {
		logScrubbedError(c, err)
	}
	_ = c.JSON(classified.status, models.GenericResponse{
		Success: false,
		Errors:  []string{classified.message},
		Code:    string(classified.code),
	})
}
