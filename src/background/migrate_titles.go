package background

import (
	"vatusa-cobalt/legacy_migration"
)

func MigrateTitles(args []string) error {
	return legacy_migration.BulkMigrateTitles()
}
