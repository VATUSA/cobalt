package endpoints

import (
	"errors"
	"net/http"
	"strconv"
	"vatusa-cobalt/acl"
	"vatusa-cobalt/dbconn"
	"vatusa-cobalt/models"
	"vatusa-cobalt/roster"

	"github.com/labstack/echo/v5"
)

func GetUser(c *echo.Context) error {
	canSeeSensitiveFields := HasGlobal(c, acl.ObjectUserSensitiveDetails, acl.ActionRead)
	cid := c.Param("cid")
	cidInt, err := strconv.Atoi(cid)
	if err != nil {
		return GenericError(c, http.StatusBadRequest, errors.New("invalid cid"))
	}

	user, err := dbconn.GetCombinedUserByCID(cidInt)
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}
	if user == nil {
		return GenericError(c, http.StatusNotFound, errors.New("user not found"))
	}

	output := models.UserFromDatabase(*user, canSeeSensitiveFields)

	return c.JSON(http.StatusOK, output)
}

func GetUserBlockers(c *echo.Context) error {
	if !AssertGlobal(c, acl.ObjectUserSensitiveDetails, acl.ActionRead) {
		return nil
	}
	cid := c.Param("cid")
	cidInt, err := strconv.Atoi(cid)
	if err != nil {
		return GenericError(c, http.StatusBadRequest, errors.New("invalid cid"))
	}
	user, err := dbconn.GetCombinedUserByCID(cidInt)
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}
	if user == nil {
		return GenericError(c, http.StatusNotFound, errors.New("user not found"))
	}
	userBlockers := roster.GetUserBlockers(*user)
	return c.JSON(http.StatusOK, userBlockers)
}
