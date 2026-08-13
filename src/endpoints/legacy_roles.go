package endpoints

import (
	"net/http"
	"vatusa-cobalt/acl"
	"vatusa-cobalt/legacy_migration"
	"vatusa-cobalt/models"

	"github.com/labstack/echo/v5"
)

func LegacySyncRoles(c *echo.Context) error {
	if !AssertGlobal(c, acl.ObjectLegacyRoleSync, acl.ActionWrite) {
		return nil
	}
	var request models.SyncRolesRequest
	err := c.Bind(&request)
	if err != nil {
		return err
	}
	err = legacy_migration.SyncRolesForUser(request)
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, "sync roles successful")
}

func LegacySyncRolesBulk(c *echo.Context) error {

	if !AssertGlobal(c, acl.ObjectLegacyRoleSync, acl.ActionWrite) {
		return nil
	}
	var request models.BulkSyncRolesRequest
	err := c.Bind(&request)
	if err != nil {
		return err
	}
	for _, req := range request.Requests {
		err = legacy_migration.SyncRolesForUser(req)
		if err != nil {
			return RespondError(c, http.StatusInternalServerError, err)
		}
	}

	return c.JSON(http.StatusOK, "sync roles successful")
}
