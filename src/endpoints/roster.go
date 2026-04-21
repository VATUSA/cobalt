package endpoints

import (
	"errors"
	"net/http"
	"vatusa-cobalt/acl"
	"vatusa-cobalt/dbconn"
	"vatusa-cobalt/models"
	"vatusa-cobalt/roster"

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

func ActionFacilityPendingTransfer(c *echo.Context) error {
	facility := c.Param("facility")
	if !AssertFacility(c, facility, acl.ObjectUserSensitiveDetails, acl.ActionWrite) {
		return nil
	}
	var request models.TransferAction
	err := c.Bind(&request)
	if err != nil {
		return GenericError(c, http.StatusBadRequest, err)
	}
	if request.Action == roster.TransferReject && request.Reason == "" {
		return GenericError(c, http.StatusBadRequest, errors.New("reason is required when rejecting requests"))
	}
	// TODO: Finish this
	return nil
}
