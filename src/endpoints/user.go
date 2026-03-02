package endpoints

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"vatusa-cobalt/acl"
	"vatusa-cobalt/models"

	"github.com/labstack/echo/v5"
)

func (h EndpointHandler) GetUser(c *echo.Context) error {
	canSeeSensitiveFields := h.HasGlobal(c, acl.ObjectUserSensitiveDetails, acl.ActionRead)
	ctx := context.Background()
	cid := c.Param("cid")
	cidInt, err := strconv.Atoi(cid)
	if err != nil {
		return GenericError(c, http.StatusBadRequest, errors.New("invalid cid"))
	}

	user, err := h.Queries.GetCombinedUserByCID(ctx, int64(cidInt))
	if err != nil {
		return GenericError(c, http.StatusNotFound, err)
	}

	output := models.UserFromDatabase(user, canSeeSensitiveFields)

	return c.JSON(http.StatusOK, output)
}
