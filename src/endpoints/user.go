package endpoints

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
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

	dbRoles, err := dbconn.Queries().GetRolesForUser(context.Background(), int32(cidInt))
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}

	output := models.UserFromDatabase(*user, canSeeSensitiveFields, models.UserRolesFromDatabase(dbRoles))

	return c.JSON(http.StatusOK, output)
}

// SearchUsers looks users up by partial cid or by name. It is gated on
// user_sensitive_details:read because matching against names lets a caller
// enumerate the user base by probing, which the redaction in UserFromDatabase
// alone would not prevent.
func SearchUsers(c *echo.Context) error {
	if !AssertGlobal(c, acl.ObjectUserSensitiveDetails, acl.ActionRead) {
		return nil
	}
	query := c.QueryParam("q")
	if strings.TrimSpace(query) == "" {
		return GenericError(c, http.StatusBadRequest, errors.New("q is required"))
	}

	limit, err := strconv.Atoi(c.QueryParam("limit"))
	if err != nil {
		limit = 5
	} else if limit < 1 {
		limit = 1
	} else if limit > 100 {
		limit = 100
	}

	users, err := dbconn.SearchUsers(query, limit)
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}

	// The gate above is the same permission that unredacts names, so results
	// here are always built with sensitive fields included.
	output := models.UsersFromDatabase(users, true)

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
