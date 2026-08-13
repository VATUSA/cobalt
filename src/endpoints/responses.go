package endpoints

import (
	"errors"
	"net/http"
	"vatusa-cobalt/models"

	"github.com/labstack/echo/v5"
)

func RespondSuccess(c *echo.Context, id int) error {
	response := models.GenericResponse{
		Success: true,
		Id:      id,
	}
	return c.JSON(http.StatusOK, response)
}

// RespondError is the choke point every error response flows through. The first
// error becomes the payload; the rest are diagnostic detail, logged but never
// serialized. Unexpected internals are scrubbed; wrap a message in SafeError to
// keep it client-visible.
func RespondError(c *echo.Context, statusCode int, errors ...error) error {
	response := models.GenericResponse{
		Success: false,
		Errors:  []string{},
	}
	status := statusCode
	for i, err := range errors {
		classified := classifyError(statusCode, err)
		if i == 0 {
			status = classified.status
			response.Code = string(classified.code)
			response.Errors = append(response.Errors, classified.message)
		}
		if classified.log {
			logScrubbedError(c, err)
		}
	}
	return c.JSON(status, response)
}

func RespondForbidden(c *echo.Context) error {
	return RespondError(c, http.StatusForbidden, errors.New("permission denied"))
}
