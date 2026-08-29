package endpoints

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"vatusa-cobalt/acl"
	"vatusa-cobalt/dbconn"

	"github.com/labstack/echo/v5"
)

func HasGlobal(c *echo.Context, object acl.Object, action acl.Action) bool {
	return acl.GetPermissionHandlerCache().HasGlobal(c, object, action)
}

func AssertGlobal(c *echo.Context, object acl.Object, action acl.Action) bool {
	if HasGlobal(c, object, action) {
		return true
	}
	_ = GenericError(c, http.StatusForbidden,
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
	_ = GenericError(c, http.StatusForbidden,
		errors.New(fmt.Sprintf("missing acl %s %s:%s", facility, object, action)))
	return false
}

func GetPermissionHandler(c *echo.Context) *acl.PermissionHandler {
	return acl.GetPermissionHandlerCache().GetHandler(c)
}

// AssertFacilityForCid checks whether the caller may perform action on object for the
// specific controller identified by targetCid — either via a global grant, or a facility
// grant matching that controller's current home or visiting facility. Returns the
// controller's current home facility (for stamping onto the record being written) and true
// on success; writes the 403/404 itself and returns false on failure.
func AssertFacilityForCid(c *echo.Context, targetCid int, object acl.Object, action acl.Action) (string, bool) {
	user, err := dbconn.GetCombinedUserByCID(targetCid)
	if err != nil || user == nil {
		_ = GenericError(c, http.StatusNotFound, errors.New("user not found"))
		return "", false
	}

	if HasGlobal(c, object, action) {
		return user.Facility, true
	}
	if HasFacility(c, user.Facility, object, action) {
		return user.Facility, true
	}
	if user.VisitingFacilities.Valid {
		for _, facility := range strings.Split(user.VisitingFacilities.String, ",") {
			if HasFacility(c, facility, object, action) {
				return user.Facility, true
			}
		}
	}
	_ = GenericError(c, http.StatusForbidden,
		errors.New(fmt.Sprintf("missing acl %s:%s for cid %d", object, action, targetCid)))
	return "", false
}
