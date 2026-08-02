package acl

import "slices"

type Role string

type ScopedRole struct {
	Role     Role
	Facility string
}

const ScopedRoleGlobalFacility = "*"

const (
	// System Administrator Role
	RoleSystemAdmin Role = "system_administrator"

	// Automatic Roles, based on state
	RoleAuthenticatedUser Role = "authenticated_user"
	RoleDivisionMember    Role = "division_member"
)

// AutomaticRoles are granted to every logged-in user based on state, not an
// explicit staff role assignment. Anything outside this set counts as staff.
var AutomaticRoles = []Role{
	RoleAuthenticatedUser,
	RoleDivisionMember,
}

func IsStaffRole(role Role) bool {
	return !slices.Contains(AutomaticRoles, role)
}

const (
	// Division Staff Roles
	RoleDivisionStaff      Role = "division_staff"
	RoleDivisionManagement Role = "division_management"
	RoleDivisionTechTeam   Role = "division_tech_team"

	// Division Teams
	RoleDivisionCommandCenter Role = "division_command_center"
	RoleSocialMediaTeam       Role = "social_media_team"
	RoleACETeam               Role = "ace_team"

	// Special role for email access
	RoleEmailUser Role = "email_user"

	// Facility Staff Roles
	RoleAirTrafficManager       Role = "air_traffic_manager"
	RoleDeputyAirTrafficManager Role = "deputy_air_traffic_manager"
	RoleTrainingAdministrator   Role = "training_administrator"
	RoleEventCoordinator        Role = "event_coordinator"
	RoleFacilityEngineer        Role = "facility_engineer"
	RoleWebMaintainer           Role = "web_maintainer"

	// Training Staff Roles
	RoleInstructor            Role = "instructor"
	RoleMentor                Role = "mentor"
	RoleDivisionAcademyEditor Role = "division_academy_editor"
	RoleFacilityAcademyEditor Role = "facility_academy_editor"

	// API Key Roles
	RoleSystemInternal Role = "system_internal"
	RoleSystemFacility Role = "system_facility"
	RoleSystemExternal Role = "system_external"
)

// RoleToPermissionObjectMap defines which Object ActionWrite
// permission should be checked to determine if the role
// can be granted or removed
var RoleToPermissionObjectMap = map[Role]Object{
	RoleSystemAdmin:           ObjectSystemRole,
	RoleDivisionStaff:         ObjectDivisionManagementRole,
	RoleDivisionManagement:    ObjectDivisionManagementRole,
	RoleDivisionTechTeam:      ObjectDivisionStaffRole,
	RoleDivisionCommandCenter: ObjectDivisionStaffRole,
	RoleSocialMediaTeam:       ObjectDivisionStaffRole,
	RoleACETeam:               ObjectDivisionStaffRole,
	RoleEmailUser:             ObjectDivisionStaffRole,

	RoleAirTrafficManager:       ObjectFacilitySeniorStaffRole,
	RoleDeputyAirTrafficManager: ObjectFacilitySeniorStaffRole,
	RoleTrainingAdministrator:   ObjectFacilitySeniorStaffRole,

	RoleEventCoordinator: ObjectFacilityJuniorStaffRole,
	RoleFacilityEngineer: ObjectFacilityJuniorStaffRole,
	RoleWebMaintainer:    ObjectFacilityJuniorStaffRole,

	RoleInstructor:            ObjectDivisionStaffRole,
	RoleMentor:                ObjectFacilityTrainingRole,
	RoleDivisionAcademyEditor: ObjectDivisionStaffRole,
	RoleFacilityAcademyEditor: ObjectFacilityTrainingRole,

	RoleSystemInternal: ObjectSystemRole,
	RoleSystemExternal: ObjectSystemRole,
	RoleSystemFacility: ObjectSystemRole,
}

// TitleCodeToPermissionObjectMap defines which Object ActionWrite
// permission should be checked to determine if a facility title with the
// given code can be granted or removed
var TitleCodeToPermissionObjectMap = map[string]Object{
	"US1":  ObjectFacilityTitleSeniorStaff,
	"US2":  ObjectFacilityTitleSeniorStaff,
	"US3":  ObjectFacilityTitleSeniorStaff,
	"US4":  ObjectFacilityTitleSeniorStaff,
	"US5":  ObjectFacilityTitleSeniorStaff,
	"US6":  ObjectFacilityTitleSeniorStaff,
	"US7":  ObjectFacilityTitleSeniorStaff,
	"US8":  ObjectFacilityTitleSeniorStaff,
	"US9":  ObjectFacilityTitleSeniorStaff,
	"ATM":  ObjectFacilityTitleSeniorStaff,
	"DATM": ObjectFacilityTitleSeniorStaff,
	"TA":   ObjectFacilityTitleSeniorStaff,

	"EC":  ObjectFacilityTitleJuniorStaff,
	"FE":  ObjectFacilityTitleJuniorStaff,
	"WM":  ObjectFacilityTitleJuniorStaff,
	"MTR": ObjectFacilityTitleJuniorStaff,
	"INS": ObjectFacilityTitleJuniorStaff,
}

// Role Scoping
var (
	GlobalRoles = []Role{
		RoleSystemAdmin,
		RoleAuthenticatedUser,
		RoleDivisionMember,
		RoleDivisionStaff,
		RoleDivisionManagement,
		RoleDivisionTechTeam,
		RoleSystemInternal,
		RoleSystemExternal,
		RoleDivisionCommandCenter,
		RoleSocialMediaTeam,
		RoleACETeam,
		RoleDivisionAcademyEditor,
		RoleEmailUser,
	}
	FacilityRoles = []Role{
		RoleAirTrafficManager,
		RoleDeputyAirTrafficManager,
		RoleTrainingAdministrator,
		RoleEventCoordinator,
		RoleFacilityEngineer,
		RoleWebMaintainer,
		RoleInstructor,
		RoleMentor,
		RoleSystemFacility,
		RoleFacilityAcademyEditor,
		RoleEmailUser,
	}

	// SystemRoles Object = ObjectSystemRole
	SystemRoles = []Role{
		RoleSystemAdmin,
		RoleSystemInternal,
		RoleSystemFacility,
		RoleSystemExternal,
	}

	// DivisionManagementRoles Object = ObjectDivisionManagementRole
	DivisionManagementRoles = []Role{
		RoleDivisionManagement,
		RoleDivisionStaff,
	}

	// DivisionStaffRoles Object = ObjectDivisionStaffRole
	DivisionStaffRoles = []Role{
		RoleInstructor,
		RoleDivisionAcademyEditor,
		RoleDivisionCommandCenter,
		RoleSocialMediaTeam,
		RoleACETeam,
		RoleDivisionTechTeam,
	}

	// FacilitySeniorStaffRoles Object = ObjectFacilitySeniorStaffRole
	FacilitySeniorStaffRoles = []Role{
		RoleAirTrafficManager,
		RoleDeputyAirTrafficManager,
		RoleTrainingAdministrator,
	}

	// FacilityJuniorStaffRoles Object = ObjectFacilityJuniorStaffRole
	FacilityJuniorStaffRoles = []Role{
		RoleEventCoordinator,
		RoleFacilityEngineer,
		RoleWebMaintainer,
	}

	// FacilityTrainingRoles Object = ObjectFacilityTrainingRole
	FacilityTrainingRoles = []Role{
		RoleMentor,
		RoleFacilityAcademyEditor,
	}
)

// Role to Permission Mapping
var (
	RoleGlobalPermissions = map[Role][]PermissionDefinition{
		RoleAuthenticatedUser: {},
		RoleDivisionMember:    {},
		RoleSystemAdmin: {
			{
				Action: ActionWrite,
				Object: ObjectSystemRole,
			},
			{
				Action: ActionWrite,
				Object: ObjectDivisionManagementRole,
			},
			{
				Action: ActionUsage,
				Object: ObjectSuperAdmin,
			},
		},
		RoleDivisionManagement: {
			{
				Action: ActionWrite,
				Object: ObjectDivisionManagementRole,
			},
			{
				Action: ActionWrite,
				Object: ObjectDivisionStaffRole,
			},
			{
				Action: ActionRead,
				Object: ObjectUserSensitiveDetails,
			},
			{
				Action: ActionWrite,
				Object: ObjectRoster,
			},
			{
				Action: ActionWrite,
				Object: ObjectFacilityTitleManagement,
			},
			{
				Action: ActionWrite,
				Object: ObjectFacilityTitleSeniorStaff,
			},
			{
				Action: ActionWrite,
				Object: ObjectFacilityTitleJuniorStaff,
			},
		},
		RoleDivisionStaff: {
			{
				Action: ActionWrite,
				Object: ObjectNewsPost,
			},
			{
				Action: ActionManageUnowned,
				Object: ObjectNewsPost,
			},
			{
				Action: ActionWrite,
				Object: ObjectEvent,
			},
			{
				Action: ActionManageUnowned,
				Object: ObjectEvent,
			},
			{
				Action: ActionRead,
				Object: ObjectEventApproval,
			},
			{
				Action: ActionWrite,
				Object: ObjectEventApproval,
			},
			{
				Action: ActionWrite,
				Object: ObjectDivisionStaffRole,
			},
			{
				Action: ActionWrite,
				Object: ObjectFacilitySeniorStaffRole,
			},
			{
				Action: ActionWrite,
				Object: ObjectFacilityJuniorStaffRole,
			},
			{
				Action: ActionRead,
				Object: ObjectUserSensitiveDetails,
			},
			{
				Action: ActionWrite,
				Object: ObjectRoster,
			},
			{
				Action: ActionRead,
				Object: ObjectFacilityTechConfig,
			},
		},
		RoleAirTrafficManager: {
			{
				Action: ActionWrite,
				Object: ObjectNewsPost,
			},
			{
				Action: ActionRead,
				Object: ObjectUserSensitiveDetails,
			},
		},
		RoleDeputyAirTrafficManager: {
			{
				Action: ActionWrite,
				Object: ObjectNewsPost,
			},
			{
				Action: ActionRead,
				Object: ObjectUserSensitiveDetails,
			},
		},
		RoleTrainingAdministrator: {
			{
				Action: ActionRead,
				Object: ObjectUserSensitiveDetails,
			},
		},

		RoleSystemInternal: {
			{
				Action: ActionWrite,
				Object: ObjectLegacyRoleSync,
			},
			{
				Action: ActionWrite,
				Object: ObjectLegacyLoginToken,
			},
			{
				Action: ActionRead,
				Object: ObjectLegacyLoginToken,
			},
			{
				Action: ActionRead,
				Object: ObjectUserSensitiveDetails,
			},
			{
				Action: ActionWrite,
				Object: ObjectRoster,
			},
		},
	}
	RoleFacilityPermissions = map[Role][]PermissionDefinition{
		RoleAirTrafficManager: {
			{
				Action: ActionWrite,
				Object: ObjectFacilityJuniorStaffRole,
			},
			{
				Action: ActionWrite,
				Object: ObjectFacilityTrainingRole,
			},
			{
				Action: ActionWrite,
				Object: ObjectEvent,
			},
			{
				Action: ActionManageUnowned,
				Object: ObjectEvent,
			},
			{
				Action: ActionWrite,
				Object: ObjectRoster,
			},
			{
				Action: ActionRead,
				Object: ObjectFacilityTechConfig,
			},
			{
				Action: ActionWrite,
				Object: ObjectFacilityTitleJuniorStaff,
			},
		},
		RoleDeputyAirTrafficManager: {
			{
				Action: ActionWrite,
				Object: ObjectFacilityJuniorStaffRole,
			},
			{
				Action: ActionWrite,
				Object: ObjectFacilityTrainingRole,
			},
			{
				Action: ActionWrite,
				Object: ObjectEvent,
			},
			{
				Action: ActionManageUnowned,
				Object: ObjectEvent,
			},
			{
				Action: ActionWrite,
				Object: ObjectRoster,
			},
			{
				Action: ActionRead,
				Object: ObjectFacilityTechConfig,
			},
			{
				Action: ActionWrite,
				Object: ObjectFacilityTitleJuniorStaff,
			},
		},
		RoleTrainingAdministrator: {
			{
				Action: ActionWrite,
				Object: ObjectFacilityTrainingRole,
			},
			{
				Action: ActionWrite,
				Object: ObjectFacilityTitleJuniorStaff,
			},
		},
		RoleEventCoordinator: {
			{
				Action: ActionWrite,
				Object: ObjectEvent,
			},
		},
		RoleFacilityEngineer: {},
		RoleWebMaintainer: {
			{
				Action: ActionRead,
				Object: ObjectFacilityTechConfig,
			},
		},
		RoleInstructor: {},
		RoleMentor:     {},

		RoleSystemFacility: {
			{
				Action: ActionRead,
				Object: ObjectUserSensitiveDetails,
			},
			{
				Action: ActionWrite,
				Object: ObjectRoster,
			},
		},
	}
)

func globalPermissionsForRole(role Role) []PermissionDefinition {
	res, ok := RoleGlobalPermissions[role]
	if !ok {
		res = []PermissionDefinition{}
	}
	return res
}

func facilityPermissionsForRole(role Role) []PermissionDefinition {
	res, ok := RoleFacilityPermissions[role]
	if !ok {
		res = []PermissionDefinition{}
	}
	return res
}
