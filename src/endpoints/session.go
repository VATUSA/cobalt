package endpoints

import (
	"net/http"
	"vatusa-cobalt/auth"
	"vatusa-cobalt/dbconn"
	"vatusa-cobalt/models"

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

	userModel := models.UserFromDatabase(*user, true)

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
			Facility: string(permission.Object),
		})
	}

	return c.JSON(http.StatusOK, session)
}
