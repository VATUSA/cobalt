package acl

import (
	"sync"
	"vatusa-cobalt/auth"
	"vatusa-cobalt/db"

	"github.com/labstack/echo/v5"
)

type PermissionHandlerCache struct {
	sync.RWMutex
	userCache map[int]*PermissionHandler
	apiCache  map[int]*PermissionHandler
	Queries   *db.Queries
}

func NewPermissionHandlerCache(queries *db.Queries) *PermissionHandlerCache {
	return &PermissionHandlerCache{
		userCache: make(map[int]*PermissionHandler),
		apiCache:  make(map[int]*PermissionHandler),
		Queries:   queries,
	}
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
		roles, _ := GetActorScopedRoles(phc.Queries, actorId)
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
		roles, _ := GetUserScopedRoles(phc.Queries, cid)
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
