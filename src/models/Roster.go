package models

import "vatusa-cobalt/db"

type Roster struct {
	Home     []User `json:"home"`
	Visitors []User `json:"visitors"`
}

func RosterFromDatabase(homeUsers []db.GetCombinedUserRow, visitUsers []db.GetCombinedUserRow, canSeeSensitiveFields bool) Roster {
	return Roster{
		Home:     UsersFromDatabase(homeUsers, canSeeSensitiveFields),
		Visitors: UsersFromDatabase(visitUsers, canSeeSensitiveFields),
	}
}
