package acl

var LegacyRoleToRolesMap = map[string][]Role{
	"US0": {
		RoleDivisionStaff,
		RoleDivisionManagement,
		RoleSystemAdmin,
	},
	"US1": {
		RoleDivisionStaff,
		RoleDivisionManagement,
	},
	"US2": {
		RoleDivisionStaff,
		RoleDivisionManagement,
	},
	"US3": {
		RoleDivisionStaff,
		RoleDivisionManagement,
	},
	"US4": {
		RoleDivisionStaff,
		RoleDivisionManagement,
	},
	"US5": {RoleDivisionStaff},
	"US6": {
		RoleDivisionStaff,
		RoleSystemAdmin,
	},
	"US7":  {RoleDivisionStaff},
	"US8":  {RoleDivisionStaff},
	"US9":  {RoleDivisionStaff},
	"USWT": {RoleDivisionTechTeam},

	"CBT":  {RoleDivisionAcademyEditor},
	"DCC":  {RoleDivisionCommandCenter},
	"DICE": {},
	"SMT":  {RoleSocialMediaTeam},
	"ACE":  {RoleACETeam},

	"EMAIL":  {RoleEmailUser},
	"FACCBT": {RoleFacilityAcademyEditor},

	"ATM":  {RoleAirTrafficManager},
	"DATM": {RoleDeputyAirTrafficManager},
	"TA":   {RoleTrainingAdministrator},
	"EC":   {RoleEventCoordinator},
	"FE":   {RoleFacilityEngineer},
	"WM":   {RoleWebMaintainer},
	"INS":  {RoleInstructor},
	"MTR":  {RoleMentor},
}

var LegacyGlobalRoles = []string{
	"US0",
	"US1",
	"US2",
	"US3",
	"US4",
	"US5",
	"US6",
	"US7",
	"US8",
	"US9",
	"USWT",
}
