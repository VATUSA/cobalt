package endpoints

import (
	"errors"
	"net/http"
	"vatusa-cobalt/auth"
	"vatusa-cobalt/dbconn"
	"vatusa-cobalt/models"
	"vatusa-cobalt/roster"

	"github.com/labstack/echo/v5"
)

func GetMySession(c *echo.Context) error {
	cid := auth.GetUserCid(c)
	if cid == -1 {
		return c.JSON(http.StatusUnauthorized, models.Session{
			User:                nil,
			GlobalPermissions:   []models.GlobalPermission{},
			FacilityPermissions: []models.FacilityPermission{},
		})
	}
	user, err := dbconn.GetCombinedUserByCID(cid)
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}
	if user == nil {
		return c.JSON(http.StatusNotFound, models.Session{})
	}

	userModel := models.UserFromDatabase(*user, true, nil)

	permissionHandler := GetPermissionHandler(c)

	globalPermissions := permissionHandler.GetGlobalPermissions()
	facilityPermissions := permissionHandler.GetFacilityPermissions()

	session := models.Session{
		User:                &userModel,
		GlobalPermissions:   []models.GlobalPermission{},
		FacilityPermissions: []models.FacilityPermission{},
	}

	for _, permission := range globalPermissions {
		session.GlobalPermissions = append(session.GlobalPermissions, models.GlobalPermission{
			Action: string(permission.Action),
			Object: string(permission.Object),
		})
	}

	for _, permission := range facilityPermissions {
		session.FacilityPermissions = append(session.FacilityPermissions, models.FacilityPermission{
			Action:   string(permission.Action),
			Object:   string(permission.Object),
			Facility: permission.Facility,
		})
	}

	return c.JSON(http.StatusOK, session)
}

func MySubmitTransferRequest(c *echo.Context) error {
	cid := auth.GetUserCid(c)
	if cid == -1 {
		return GenericError(c, http.StatusUnauthorized, errors.New("login required"))
	}
	user, err := dbconn.GetCombinedUserByCID(cid)
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}
	if user == nil {
		return GenericError(c, http.StatusNotFound, errors.New("user not found"))
	}
	var request models.MyTransferRequestRequest
	err = c.Bind(&request)
	if err != nil {
		return GenericError(c, http.StatusBadRequest, err)
	}
	blockers := roster.GetUserBlockers(*user)
	if blockers.IsTransferBlocked {
		return GenericError(c, http.StatusBadRequest, errors.New("user is transfer blocked"))
	}
	rec, err := roster.CreateTransferRequest(*user, request.ToFacility, request.Reason, *user)
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}
	return GenericSuccess(c, int(rec.ID))
}
