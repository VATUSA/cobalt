package acl

import (
	"sync"
	"vatusa-cobalt/auth"

	"github.com/labstack/echo/v5"
)

type PermissionHandlerCache struct {
	sync.RWMutex
	userCache map[int]*PermissionHandler
	apiCache  map[int]*PermissionHandler
}

func newPermissionHandlerCache() *PermissionHandlerCache {
	return &PermissionHandlerCache{
		userCache: make(map[int]*PermissionHandler),
		apiCache:  make(map[int]*PermissionHandler),
	}
}

var permissionHandlerCache *PermissionHandlerCache = newPermissionHandlerCache()

func GetPermissionHandlerCache() *PermissionHandlerCache {
	return permissionHandlerCache
}

func (phc *PermissionHandlerCache) getPermissionHandlerFromContext(c *echo.Context) *PermissionHandler {
	actorId := auth.GetActorId(c)
	if actorId > 0 {
		phc.RLock()
		res, ok := phc.apiCache[actorId]
		phc.RUnlock()
		if ok && !res.IsStale() {
			return res
		}
		roles, _ := GetActorScopedRoles(actorId)
		ph := NewPermissionHandler(roles)
		phc.Lock()
		phc.apiCache[actorId] = ph
		phc.Unlock()
		return ph
	}
	if auth.IsLoggedIn(c) {
		cid := auth.GetUserCid(c)
		phc.RLock()
		res, ok := phc.userCache[cid]
		phc.RUnlock()
		if ok && !res.IsStale() {
			return res
		}
		roles, _ := GetUserScopedRoles(cid)
		ph := NewPermissionHandler(roles)
		phc.Lock()
		phc.userCache[cid] = ph
		phc.Unlock()
		return ph
	}
	return NewPermissionHandler([]ScopedRole{})
}

func (phc *PermissionHandlerCache) HasGlobal(c *echo.Context, object Object, action Action) bool {
	ph := phc.getPermissionHandlerFromContext(c)
	return ph.HasGlobal(object, action)
}

func (phc *PermissionHandlerCache) HasFacility(c *echo.Context, facility string, object Object, action Action) bool {
	ph := phc.getPermissionHandlerFromContext(c)
	return ph.HasFacility(facility, object, action)
}

func (phc *PermissionHandlerCache) GetHandler(c *echo.Context) *PermissionHandler {
	ph := phc.getPermissionHandlerFromContext(c)
	return ph
}

func (phc *PermissionHandlerCache) GetHandlerForCid(cid int) *PermissionHandler {
	phc.RLock()
	res, ok := phc.userCache[cid]
	phc.RUnlock()
	if ok && !res.IsStale() {
		return res
	}
	roles, _ := GetUserScopedRoles(cid)
	ph := NewPermissionHandler(roles)
	phc.Lock()
	phc.userCache[cid] = ph
	phc.Unlock()
	return ph
}
