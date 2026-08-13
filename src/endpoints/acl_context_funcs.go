package endpoints

import (
	"errors"
	"fmt"
	"net/http"
	"vatusa-cobalt/acl"

	"github.com/labstack/echo/v5"
)

func HasGlobal(c *echo.Context, object acl.Object, action acl.Action) bool {
	return acl.GetPermissionHandlerCache().HasGlobal(c, object, action)
}

func AssertGlobal(c *echo.Context, object acl.Object, action acl.Action) bool {
	if HasGlobal(c, object, action) {
		return true
	}
	_ = RespondError(c, http.StatusForbidden,
		errors.New(fmt.Sprintf("missing acl global %s:%s", object, action)))
	return false
}

func HasFacility(c *echo.Context, facility string, object acl.Object, action acl.Action) bool {
	return acl.GetPermissionHandlerCache().HasFacility(c, facility, object, action)
}

func AssertFacility(c *echo.Context, facility string, object acl.Object, action acl.Action) bool {
	if HasFacility(c, facility, object, action) {
		return true
	}
	_ = RespondError(c, http.StatusForbidden,
		errors.New(fmt.Sprintf("missing acl %s %s:%s", facility, object, action)))
	return false
}

func GetPermissionHandler(c *echo.Context) *acl.PermissionHandler {
	return acl.GetPermissionHandlerCache().GetHandler(c)
}
