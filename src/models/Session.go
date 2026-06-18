package models

type Session struct {
	User                *User                `json:"user"`
	GlobalPermissions   []GlobalPermission   `json:"global_permissions"`
	FacilityPermissions []FacilityPermission `json:"facility_permissions"`
}

type GlobalPermission struct {
	Action string `json:"action"`
	Object string `json:"object"`
}

type FacilityPermission struct {
	Action   string `json:"action"`
	Object   string `json:"object"`
	Facility string `json:"facility"`
}

type TokenSessionRequest struct {
	Token string `json:"token"`
}
