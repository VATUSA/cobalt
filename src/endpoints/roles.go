package endpoints

import (
	"net/http"
	"vatusa-cobalt/acl"
	"vatusa-cobalt/dbconn"
	"vatusa-cobalt/models"

	"github.com/labstack/echo/v5"
)

func GetACETeam(c *echo.Context) error {
	users, err := dbconn.GetUsersByRole(string(acl.RoleACETeam), nil)
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}
	output := models.ACETeamMembersFromDatabase(users)
	return c.JSON(http.StatusOK, output)
}
