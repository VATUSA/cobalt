package auth

import (
	"context"
	"log"
	"sync"
	"time"
	"vatusa-cobalt/db"
)

type UserPermission = string

// All Permissions
const (
	PermPostNews     UserPermission = "post_news"
	PermManageEvents UserPermission = "manage_events"
)

type UserRole = string

// All Roles
const (
	RoleAirTrafficManager       UserRole = "ATM"
	RoleDeputyAirTrafficManager UserRole = "DATM"
	RoleTrainingAdministrator   UserRole = "TA"
	RoleEventCoordinator        UserRole = "EC"
	RoleFacilityEngineer        UserRole = "FE"
	RoleWebMaintainer           UserRole = "WM"
	RoleVATUSAStaff             UserRole = "USA"
	RoleVATUSAManagement        UserRole = "MGMT"
	RoleVATUSAWebTeam           UserRole = "USWT"
	RoleACETeam                 UserRole = "ACE"
	RoleAcademyEditor           UserRole = "CBT"
	RoleFacilityAcademyEditor   UserRole = "FACCBT"
	RoleInstructor              UserRole = "INS"
	RoleMentor                  UserRole = "MTR"
	RoleSocialMediaTeam         UserRole = "SMT"
	RoleCommandCenterStaff      UserRole = "DCC"
)

var (
	GlobalRoles = []UserRole{
		RoleVATUSAStaff,
		RoleVATUSAManagement,
		RoleVATUSAWebTeam,
		RoleACETeam,
		RoleAcademyEditor,
		RoleSocialMediaTeam,
		RoleCommandCenterStaff,
	}
	FacilityRoles = []UserRole{
		RoleAirTrafficManager,
		RoleDeputyAirTrafficManager,
		RoleTrainingAdministrator,
		RoleEventCoordinator,
		RoleFacilityEngineer,
		RoleWebMaintainer,
		RoleFacilityAcademyEditor,
		RoleInstructor,
		RoleMentor,
	}
	RoleGlobalPermissions = map[UserRole][]UserPermission{
		RoleVATUSAStaff: []UserPermission{
			PermPostNews,
			PermManageEvents,
		},
		RoleVATUSAManagement: []UserPermission{
			PermPostNews,
			PermManageEvents,
		},
		RoleAirTrafficManager: []UserPermission{
			PermPostNews,
		},
		RoleDeputyAirTrafficManager: []UserPermission{
			PermPostNews,
		},
	}
	RoleFacilityPermissions = map[UserRole][]UserPermission{
		RoleAirTrafficManager: []UserPermission{
			PermManageEvents,
		},
		RoleDeputyAirTrafficManager: []UserPermission{
			PermManageEvents,
		},
		RoleEventCoordinator: []UserPermission{
			PermManageEvents,
		},
	}
)

type FacilityRole struct {
	Facility string
	Role     UserRole
}

type UserPermissionHandler struct {
	sync.RWMutex
	cid                   int
	globalPermissionMap   map[UserPermission]bool
	facilityPermissionMap map[string]map[UserPermission]bool
	loadedAt              time.Time
	queries               *db.Queries
}

func (uph *UserPermissionHandler) HasGlobalPermission(permission UserPermission) bool {
	uph.RLock()
	result, ok := uph.globalPermissionMap[permission]
	uph.RUnlock()
	if !ok {
		return false
	}
	return result
}

func (uph *UserPermissionHandler) HasFacilityPermission(permission UserPermission, facility string) bool {
	if uph.HasGlobalPermission(permission) {
		return true
	}
	uph.RLock()
	nestedMap, ok := uph.facilityPermissionMap[facility]
	uph.RUnlock()
	if !ok {
		return false
	}
	uph.RLock()
	result, ok := nestedMap[permission]
	uph.RUnlock()
	if !ok {
		return false
	}
	return result
}

func (uph *UserPermissionHandler) Load() {
	uph.Lock()
	uph.globalPermissionMap = map[UserPermission]bool{}
	uph.facilityPermissionMap = map[string]map[UserPermission]bool{}
	uph.loadedAt = time.Unix(0, 0)
	uph.Unlock()
	ctx := context.Background()
	globalRolesRecords, err := uph.queries.FetchGlobalRolesByCID(ctx, int32(uph.cid))
	if err != nil {
		log.Printf("failed to load global roles for user %d: %v", uph.cid, err)
		return
	}
	var globalRoles []UserRole
	for _, globalRole := range globalRolesRecords {
		globalRoles = append(globalRoles, globalRole.Role)
	}
	facilityRolesRecords, err := uph.queries.FetchFacilityRolesByCID(ctx, int32(uph.cid))
	if err != nil {
		log.Printf("failed to load facilities roles for user %d: %v", uph.cid, err)
		uph.facilityPermissionMap = map[string]map[UserPermission]bool{}
		return
	}
	var facilityRoles []FacilityRole
	for _, fr := range facilityRolesRecords {
		facilityRoles = append(facilityRoles, FacilityRole{
			Facility: fr.Facility,
			Role:     fr.Role,
		})
	}
	uph.setFromRoles(globalRoles, facilityRoles)
	uph.Lock()
	uph.loadedAt = time.Now()
	uph.Unlock()
}

func (uph *UserPermissionHandler) setFromRoles(globalRoles []UserRole, facilityRoles []FacilityRole) {
	uph.Lock()
	for _, role := range globalRoles {
		for _, permission := range RoleGlobalPermissions[role] {
			uph.globalPermissionMap[permission] = true
		}
	}

	for _, fr := range facilityRoles {
		for _, permission := range RoleGlobalPermissions[fr.Role] {
			uph.globalPermissionMap[permission] = true
		}
		if _, ok := uph.facilityPermissionMap[fr.Facility]; !ok {
			uph.facilityPermissionMap[fr.Facility] = map[UserPermission]bool{}
		}
		for _, permission := range RoleFacilityPermissions[fr.Role] {
			uph.facilityPermissionMap[fr.Facility][permission] = true
		}
	}
	uph.Unlock()
}

func NewUserPermissionHandler(queries *db.Queries, cid int) *UserPermissionHandler {
	h := &UserPermissionHandler{
		cid:                   cid,
		globalPermissionMap:   make(map[UserPermission]bool),
		facilityPermissionMap: make(map[string]map[UserPermission]bool),
		loadedAt:              time.Unix(0, 0),
		queries:               queries,
	}
	h.Load()
	return h
}

type PermissionCache struct {
	sync.RWMutex
	userPermissionMap map[int]*UserPermissionHandler
	Queries           *db.Queries
}

func NewPermissionCache(queries *db.Queries) *PermissionCache {
	return &PermissionCache{
		userPermissionMap: map[int]*UserPermissionHandler{},
		Queries:           queries,
	}
}

func (c *PermissionCache) Flush() {
	c.Lock()
	c.userPermissionMap = make(map[int]*UserPermissionHandler)
	c.Unlock()
}

func (c *PermissionCache) Drop(cid int) {
	c.Lock()
	delete(c.userPermissionMap, cid)
	c.Unlock()
}

func (c *PermissionCache) Get(cid int) *UserPermissionHandler {
	c.RLock()
	permission, ok := c.userPermissionMap[cid]
	c.RUnlock()
	if !ok {
		permission = NewUserPermissionHandler(c.Queries, cid)
		c.Lock()
		c.userPermissionMap[cid] = permission
		c.Unlock()
	}
	return permission
}
