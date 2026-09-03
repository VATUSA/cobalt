package background

import (
	"log"
	"slices"
	"vatusa-cobalt/legacy_migration"
)

// MigratePolicies copies `vatusa-old`.policy_categories/policies into
// Cobalt's policy_category/policy_document tables. Re-runnable: see
// legacy_migration.BulkMigratePolicies for the idempotency/two-writer-guard
// design. Accepts "--dry-run" and "--force" in either order.
func MigratePolicies(args []string) error {
	opts := legacy_migration.PolicyMigrationOptions{
		DryRun: slices.Contains(args, "--dry-run"),
		Force:  slices.Contains(args, "--force"),
	}

	report, err := legacy_migration.BulkMigratePolicies(opts)
	if err != nil {
		return err
	}

	log.Printf(
		"policy migration complete (dry_run=%v force=%v): categories inserted=%d skipped=%d; documents inserted=%d updated=%d skipped=%d",
		opts.DryRun, opts.Force,
		report.CategoriesInserted, report.CategoriesSkipped,
		report.DocumentsInserted, report.DocumentsUpdated, report.DocumentsSkipped,
	)
	return nil
}
