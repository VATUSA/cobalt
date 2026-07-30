package background

import (
	"vatusa-cobalt/legacy_migration"
)

func MigrateActionLogs(args []string) error {
	return legacy_migration.BulkMigrateActionLogs()
}
