package background

import (
	"vatusa-cobalt/legacy_migration"
)

func MigrateUsers(args []string) error {
	err := legacy_migration.BulkMigrateUsers()
	if err != nil {
		return err
	}
	err = legacy_migration.BulkMigrateRoles()
	if err != nil {
		return err
	}
	return nil
}
