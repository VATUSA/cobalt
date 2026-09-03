package background

type JobFunc func(args []string) error

var JobDefinitions = map[string]JobFunc{
	"test":                TestJob,
	"vatsim_sync":         VatsimSync,
	"migrate_users":       MigrateUsers,
	"migrate_titles":      MigrateTitles,
	"check_rating_hours":  CheckRatingHours,
	"migrate_action_logs": MigrateActionLogs,
	"migrate_policies":    MigratePolicies,
}
