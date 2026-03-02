package endpoints

import (
	"context"
	"net/http"
	"vatusa-cobalt/auth"
	"vatusa-cobalt/models"

	"github.com/labstack/echo/v5"
)

func (h EndpointHandler) GetMySession(c *echo.Context) error {
	cid := auth.GetUserCid(c)
	if cid == -1 {
		return c.JSON(http.StatusOK, models.Session{
			User:                nil,
			GlobalPermissions:   []models.GlobalPermission{},
			FacilityPermissions: []models.FacilityPermission{},
		})
	}
	ctx := context.Background()
	user, err := h.Queries.GetCombinedUserByCID(ctx, int64(cid))
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}

	userModel := models.UserFromDatabase(user, true)

	permissionHandler := h.GetPermissionHandler(c)

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
