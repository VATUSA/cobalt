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

	ObjectNewsPost Object = "news_post"
	ObjectEvent    Object = "event"

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

type PermissionDefinition struct {
	Action Action
	Object Object
}
