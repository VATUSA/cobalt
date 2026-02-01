package endpoints

import (
	"vatusa-cobalt/auth"
	"vatusa-cobalt/db"

	"github.com/labstack/echo/v5"
)

type EndpointHandler struct {
	Queries         *db.Queries
	PermissionCache *auth.PermissionCache
}

func (h EndpointHandler) HasGlobalPermission(c *echo.Context, permission auth.UserPermission) bool {
	if !auth.IsLoggedIn(c) {
		return false
	}
	cid := auth.GetUserCid(c)
	upc := h.PermissionCache.Get(cid)
	return upc.HasGlobalPermission(permission)
}

func (h EndpointHandler) HasFacilityPermission(c *echo.Context, permission auth.UserPermission, facility string) bool {
	if !auth.IsLoggedIn(c) {
		return false
	}
	cid := auth.GetUserCid(c)
	upc := h.PermissionCache.Get(cid)
	return upc.HasFacilityPermission(permission, facility)
}

func NewEndpointHandler(queries *db.Queries) *EndpointHandler {
	return &EndpointHandler{
		Queries:         queries,
		PermissionCache: auth.NewPermissionCache(queries),
	}
}
