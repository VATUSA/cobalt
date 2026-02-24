package models

type BulkSyncRolesRequest struct {
	Requests []SyncRolesRequest `json:"requests"`
}

type SyncRolesRequest struct {
	CID   int          `json:"cid"`
	Roles []LegacyRole `json:"roles"`
}

type LegacyRole struct {
	Role     string `json:"role"`
	Facility string `json:"facility"`
}
