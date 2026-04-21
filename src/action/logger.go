package action

import "vatusa-cobalt/db"

func Log(subject db.GetCombinedUserRow, action Action, message string, actorCid int64) error {
	return nil
}
