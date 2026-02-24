package endpoints

import (
	"errors"
	"fmt"
	"net/http"
	"vatusa-cobalt/acl"
	"vatusa-cobalt/db"

	"github.com/labstack/echo/v5"
)

type EndpointHandler struct {
	Queries                *db.Queries
	PermissionHandlerCache *acl.PermissionHandlerCache
}

func (h EndpointHandler) HasGlobal(c *echo.Context, object acl.Object, action acl.Action) bool {
	return h.PermissionHandlerCache.HasGlobal(c, object, action)
}

func (h EndpointHandler) AssertGlobal(c *echo.Context, object acl.Object, action acl.Action) bool {
	if h.HasGlobal(c, object, action) {
		return true
	}
	_ = GenericError(c, http.StatusForbidden,
		errors.New(fmt.Sprintf("missing acl global %s:%s", object, action)))
	return false
}

func (h EndpointHandler) HasFacility(c *echo.Context, facility string, object acl.Object, action acl.Action) bool {
	return h.PermissionHandlerCache.HasFacility(c, facility, object, action)
}

func (h EndpointHandler) AssertFacility(c *echo.Context, facility string, object acl.Object, action acl.Action) bool {
	if h.HasFacility(c, facility, object, action) {
		return true
	}
	_ = GenericError(c, http.StatusForbidden,
		errors.New(fmt.Sprintf("missing acl %s %s:%s", facility, object, action)))
	return false
}

func NewEndpointHandler(queries *db.Queries) *EndpointHandler {
	return &EndpointHandler{
		Queries:                queries,
		PermissionHandlerCache: acl.NewPermissionHandlerCache(queries),
	}
}
