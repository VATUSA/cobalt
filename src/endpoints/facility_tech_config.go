package endpoints

import (
	"context"
	"database/sql"
	"net/http"
	"vatusa-cobalt/acl"
	"vatusa-cobalt/dbconn"
	"vatusa-cobalt/models"

	"github.com/labstack/echo/v5"
)

func GetFacilityApiKeys(c *echo.Context) error {
	facility := c.Param("facility")
	if !AssertFacility(c, facility, acl.ObjectFacilityTechConfig, acl.ActionRead) {
		return nil
	}

	rows, err := dbconn.Queries().GetV3ApiKeysByFacility(context.Background(), sql.NullString{String: facility, Valid: true})
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, models.V3ApiKeysFromDatabase(rows))
}
