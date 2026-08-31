package acl

type Action string

const (
	ActionRead          Action = "read"
	ActionWrite         Action = "write"
	ActionManageUnowned Action = "manage_unowned"
	ActionUsage         Action = "usage"
)

type Object string

const (
	// ObjectSuperAdmin with ActionUsage grants all permissions
	ObjectSuperAdmin Object = "superadmin"

	ObjectNewsPost             Object = "news_post"
	ObjectFaq                  Object = "faq"
	ObjectSoloCert             Object = "solo_cert"
	ObjectPolicy               Object = "policy"
	ObjectEvent                Object = "event"
	ObjectEventApproval        Object = "event_approval"
	ObjectUserSensitiveDetails Object = "user_sensitive_details"
	ObjectRoster               Object = "roster"
	ObjectFacilityTechConfig   Object = "facility_tech_config"
	ObjectFacilityTitle        Object = "facility_title"

	ObjectFacilityTitleManagement  Object = "facility_title_management"
	ObjectFacilityTitleSeniorStaff Object = "facility_title_senior_staff"
	ObjectFacilityTitleJuniorStaff Object = "facility_title_junior_staff"

	// Role Objects
	ObjectSystemRole              Object = "system_api_role"
	ObjectDivisionManagementRole  Object = "division_management_role"
	ObjectDivisionStaffRole       Object = "division_staff_role"
	ObjectFacilitySeniorStaffRole Object = "facility_senior_staff_role"
	ObjectFacilityJuniorStaffRole Object = "facility_junior_staff_role"
	ObjectFacilityTrainingRole    Object = "facility_training_role"

	// ObjectLegacyRoleSync is used for the legacy role sync process, which can write any role
	ObjectLegacyRoleSync   Object = "legacy_role_sync"
	ObjectLegacyLoginToken Object = "legacy_login_token"
)

var MaskedObjects = []Object{
	ObjectSystemRole,
	ObjectDivisionManagementRole,
	ObjectDivisionStaffRole,
	ObjectFacilitySeniorStaffRole,
	ObjectFacilityJuniorStaffRole,
	ObjectFacilityTrainingRole,

	ObjectLegacyRoleSync,
	ObjectLegacyLoginToken,
}

type PermissionDefinition struct {
	Action Action
	Object Object
}

type ScopedPermissionDefinition struct {
	Action   Action
	Object   Object
	Facility string
}
