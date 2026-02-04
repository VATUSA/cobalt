package endpoints

import (
	"errors"
	"net/http"
	"vatusa-cobalt/models"

	"github.com/labstack/echo/v5"
)

func GenericSuccess(c *echo.Context, id int) error {
	response := models.GenericResponse{
		Success: true,
		Id:      id,
	}
	return c.JSON(http.StatusOK, response)
}

func GenericError(c *echo.Context, statusCode int, errors ...error) error {
	response := models.GenericResponse{
		Success: false,
		Errors:  []string{},
	}
	for _, err := range errors {
		response.Errors = append(response.Errors, err.Error())
	}
	return c.JSON(statusCode, response)
}

func ErrorNoPermission(c *echo.Context) error {
	return GenericError(c, http.StatusForbidden, errors.New("permission denied"))
}
