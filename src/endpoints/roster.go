package endpoints

import (
	"net/http"
	"vatusa-cobalt/acl"
	"vatusa-cobalt/dbconn"
	"vatusa-cobalt/models"

	"github.com/labstack/echo/v5"
)

func GetFacilityRoster(c *echo.Context) error {
	canSeeSensitiveFields := HasGlobal(c, acl.ObjectUserSensitiveDetails, acl.ActionRead)
	facility := c.Param("facility")

	homeUsers, err := dbconn.GetCombinedUsersByHomeFacility(facility)
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}
	visitUsers, err := dbconn.GetCombinedUsersByVisitingFacility(facility)
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}

	model := models.RosterFromDatabase(homeUsers, visitUsers, canSeeSensitiveFields)
	return c.JSON(http.StatusOK, model)
}

func GetFacilityPendingTransfers(c *echo.Context) error {
	canSeeSensitiveFields := HasGlobal(c, acl.ObjectUserSensitiveDetails, acl.ActionRead)
	facility := c.Param("facility")

	transferRequests, err := dbconn.GetFacilityPendingTransferRequests(facility)
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}
	model := models.TransferRequestsCombinedFromDatabase(transferRequests, canSeeSensitiveFields)
	return c.JSON(http.StatusOK, model)
}
