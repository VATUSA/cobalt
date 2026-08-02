package action

type Action = string

const (
	TransferRequest          Action = "transfer_request"
	TransferApproved         Action = "transfer_approved"
	TransferDenied           Action = "transfer_denied"
	ForceTransfer            Action = "force_transfer"
	VisitRequest             Action = "visit_request"
	VisitApproved            Action = "visit_approved"
	VisitDenied              Action = "visit_denied"
	ForceVisit               Action = "force_visit"
	RemovedFromHomeFacility  Action = "removed_from_home_facility"
	RemovedFromVisitFacility Action = "removed_from_visit_facility"
	RoleGranted              Action = "role_granted"
	RoleRevoked              Action = "role_revoked"
	TitleGranted             Action = "title_granted"
	TitleRevoked             Action = "title_revoked"
	Migrated                 Action = "migrated"
)
