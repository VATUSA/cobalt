package endpoints

import (
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
	_ = RespondError(c, http.StatusForbidden,
		fmt.Errorf("missing acl global %s:%s", object, action))
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
		fmt.Errorf("missing acl %s %s:%s", facility, object, action))
	return false
}

// HasAny reports whether the caller holds object:action at any scope at all,
// without regard to any specific target.
func HasAny(c *echo.Context, object acl.Object, action acl.Action) bool {
	return acl.GetPermissionHandlerCache().HasAny(c, object, action)
}

func GetPermissionHandler(c *echo.Context) *acl.PermissionHandler {
	return acl.GetPermissionHandlerCache().GetHandler(c)
}

// AssertFacilityForCid checks whether the caller may perform action on object for the
// specific controller identified by targetCid — either via a global grant, or a facility
// grant matching that controller's current home or visiting facility. Returns the
// specific facility that granted access (for stamping onto the record being written) and
// true on success; writes the 403/404 itself and returns false on failure.
//
// The HasAny check runs before the GetCombinedUserByCID lookup so that a caller who could
// never perform this action at any scope gets a flat 403 without a DB round trip — and,
// more importantly, without letting the presence/absence of a 404 reveal whether targetCid
// is a real controller to a caller who holds no relevant permission at all.
func AssertFacilityForCid(c *echo.Context, targetCid int, object acl.Object, action acl.Action) (string, bool) {
	if HasGlobal(c, object, action) {
		user, err := dbconn.GetCombinedUserByCID(targetCid)
		if err != nil || user == nil {
			_ = RespondError(c, http.StatusNotFound, fmt.Errorf("user not found"))
			return "", false
		}
		return user.Facility, true
	}
	if !HasAny(c, object, action) {
		_ = RespondError(c, http.StatusForbidden,
			fmt.Errorf("missing acl %s:%s for cid %d", object, action, targetCid))
		return "", false
	}

	user, err := dbconn.GetCombinedUserByCID(targetCid)
	if err != nil || user == nil {
		_ = RespondError(c, http.StatusNotFound, fmt.Errorf("user not found"))
		return "", false
	}

	if HasFacility(c, user.Facility, object, action) {
		return user.Facility, true
	}
	if user.VisitingFacilities.Valid {
		for _, facility := range strings.Split(user.VisitingFacilities.String, ",") {
			facility = strings.TrimSpace(facility)
			if HasFacility(c, facility, object, action) {
				return facility, true
			}
		}
	}
	_ = RespondError(c, http.StatusForbidden,
		fmt.Errorf("missing acl %s:%s for cid %d", object, action, targetCid))
	return "", false
}
