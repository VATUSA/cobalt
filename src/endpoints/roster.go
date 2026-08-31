package endpoints

import (
	"context"
	"errors"
	"net/http"
	"vatusa-cobalt/acl"
	"vatusa-cobalt/auth"
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
		return RespondError(c, http.StatusInternalServerError, err)
	}
	visitUsers, err := dbconn.GetCombinedUsersByVisitingFacility(facility)
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, err)
	}

	model := models.RosterFromDatabase(homeUsers, visitUsers, canSeeSensitiveFields)
	return c.JSON(http.StatusOK, model)
}

func GetFacilityStaff(c *echo.Context) error {
	facility := c.Param("facility")

	roles := make([]string, 0, len(acl.FacilitySeniorStaffRoles)+len(acl.FacilityJuniorStaffRoles))
	for _, r := range acl.FacilitySeniorStaffRoles {
		roles = append(roles, string(r))
	}
	for _, r := range acl.FacilityJuniorStaffRoles {
		roles = append(roles, string(r))
	}

	rows, err := dbconn.GetFacilityStaffRoles(facility, roles)
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, err)
	}

	model := models.FacilityStaffFromDatabase(rows)
	return c.JSON(http.StatusOK, model)
}

func GetFacilityPendingTransfers(c *echo.Context) error {
	facility := c.Param("facility")
	if !AssertFacility(c, facility, acl.ObjectUserSensitiveDetails, acl.ActionRead) {
		return nil
	}
	canSeeSensitiveFields := HasGlobal(c, acl.ObjectUserSensitiveDetails, acl.ActionRead)

	transferRequests, err := dbconn.GetFacilityPendingTransferRequests(facility)
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, err)
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
		return RespondError(c, http.StatusBadRequest, err)
	}
	if request.Action == roster.TransferReject && request.Reason == "" {
		return RespondError(c, http.StatusBadRequest, errors.New("reason is required when rejecting requests"))
	}
	cid := auth.GetUserCid(c)
	if cid == -1 {
		// If this is an API request, check permissions for the provided CID to make sure they are authorized
		cid = int(request.ActorCid)
		ph := acl.GetPermissionHandlerCache().GetHandlerForCid(cid)
		if !ph.HasFacility(facility, acl.ObjectUserSensitiveDetails, acl.ActionWrite) {
			return RespondForbidden(c)
		}
	}
	ctx := context.Background()
	transferRequest, err := dbconn.Queries().GetTransferRequestById(ctx, request.Id)
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, err)
	}

	if request.Action == roster.TransferAccept {
		err = roster.AcceptTransferRequest(transferRequest, int64(cid))
		if err != nil {
			return respondTransferActionError(c, err)
		}
		return RespondSuccess(c, int(transferRequest.ID))
	} else if request.Action == roster.TransferReject {
		err = roster.RejectTransferRequest(transferRequest, int64(cid))
		if err != nil {
			return respondTransferActionError(c, err)
		}
		return RespondSuccess(c, int(transferRequest.ID))
	}

	return RespondError(c, http.StatusBadRequest, errors.New("invalid action"))
}

func respondTransferActionError(c *echo.Context, err error) error {
	if errors.Is(err, roster.ErrUserNotFound) {
		return RespondError(c, http.StatusNotFound, err)
	}
	return RespondError(c, http.StatusInternalServerError, err)
}
