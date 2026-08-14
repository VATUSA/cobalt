package endpoints

import (
	"errors"
	"net/http"
	"strings"
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
		return RespondError(c, http.StatusInternalServerError, err)
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

func GetMyAssignableRoles(c *echo.Context) error {
	cid := auth.GetUserCid(c)
	if cid == -1 {
		return RespondError(c, http.StatusUnauthorized, errors.New("login required"))
	}

	permissionHandler := GetPermissionHandler(c)

	roles := map[string][]string{}
	for facility, roleList := range permissionHandler.GetAssignableRoles() {
		names := make([]string, 0, len(roleList))
		for _, role := range roleList {
			names = append(names, string(role))
		}
		roles[facility] = names
	}

	return c.JSON(http.StatusOK, models.AssignableRoles{Roles: roles})
}

func MySubmitTransferRequest(c *echo.Context) error {
	cid := auth.GetUserCid(c)
	if cid == -1 {
		return RespondError(c, http.StatusUnauthorized, errors.New("login required"))
	}
	user, err := dbconn.GetCombinedUserByCID(cid)
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, err)
	}
	if user == nil {
		return RespondError(c, http.StatusNotFound, errors.New("user not found"))
	}
	var request models.MyTransferRequestRequest
	err = c.Bind(&request)
	if err != nil {
		return RespondError(c, http.StatusBadRequest, err)
	}
	blockers := roster.GetUserBlockers(*user)
	if blockers.IsTransferBlocked {
		return RespondError(c, http.StatusBadRequest, errors.New("user is transfer blocked"))
	}
	if strings.EqualFold(user.Facility, request.ToFacility) {
		return RespondError(c, http.StatusBadRequest, errors.New("from facility and to facility must not be equal"))
	}
	rec, err := roster.CreateTransferRequest(*user, request.ToFacility, request.Reason, *user)
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, err)
	}
	return RespondSuccess(c, int(rec.ID))
}
