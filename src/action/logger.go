package action

import (
	"context"
	"time"
	"vatusa-cobalt/db"
)

func Log(q *db.Queries, subject db.GetCombinedUserRow, action Action, message string, actorCid int64) error {
	now := time.Now().Unix()
	return q.CreateActionLog(context.Background(), db.CreateActionLogParams{
		ActorCid:  int32(actorCid),
		TargetCid: int32(subject.Cid),
		Action:    string(action),
		Log:       message,
		CreatedAt: now,
		UpdatedAt: now,
	})
}
