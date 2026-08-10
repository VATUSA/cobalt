package legacy_migration

import (
	"context"
	"vatusa-cobalt/action"
	"vatusa-cobalt/db"
	"vatusa-cobalt/dbconn"
)

func BulkMigrateActionLogs() error {
	ctx := context.Background()

	tx, err := dbconn.DB().BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	queries := dbconn.Queries().WithTx(tx)

	logs, err := queries.GetLegacyActionLogs(ctx)
	if err != nil {
		return err
	}

	for _, l := range logs {
		err = queries.CreateActionLog(ctx, db.CreateActionLogParams{
			ActorCid:  l.From,
			TargetCid: l.To,
			Action:    string(action.Migrated),
			Log:       l.Log,
			CreatedAt: l.CreatedAt.Unix(),
			UpdatedAt: l.UpdatedAt.Unix(),
		})
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
