package background

import "vatusa-cobalt/user_migration"

func MigrateUsers(args []string) error {
	return user_migration.BulkMigrateUsers()
}
