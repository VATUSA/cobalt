package legacy_migration

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"html"
	"log"
	"strings"
	"time"
	"vatusa-cobalt/db"
	"vatusa-cobalt/dbconn"
)

// migrationActorCid is stamped as created_by_cid/updated_by_cid on every
// document this job writes -- the same CID-0 "Automated" convention used for
// action_log rows written by other automated jobs (see roster's
// resolveActorDisplayName). It's also how the two-writer guard recognizes a
// row it's still allowed to touch: once a real staffer edits a migrated
// document, UpdatedByCid moves off this sentinel and the job leaves it alone.
const migrationActorCid = int32(0)

// legacyDocsBaseURL is the CDN origin the legacy site has always served
// policy files from (current/resources/views/info/policies.blade.php). It is
// deliberately independent of config.DocsPublicBaseURL(), which governs
// *new* uploads made through the Cobalt staff UI -- these 31 files already
// exist at this fixed location and are not being re-uploaded, so the
// migration must not depend on Spaces env config being present to run.
const legacyDocsBaseURL = "https://vatusa-storage.nyc3.cdn.digitaloceanspaces.com"

// PolicyMigrationOptions controls how BulkMigratePolicies resolves conflicts
// with rows a previous run (or a staffer, post-cutover) has already touched.
type PolicyMigrationOptions struct {
	// DryRun computes and logs what would happen without writing anything.
	DryRun bool
	// Force overwrites a policy_document whose UpdatedByCid is no longer the
	// migration sentinel, i.e. one a staffer has since edited. Categories
	// are never overwritten by Force -- see the comment on
	// migrateCategories.
	Force bool
}

// PolicyMigrationReport summarizes what BulkMigratePolicies did (or, under
// DryRun, would do), so the caller can log or assert on it.
type PolicyMigrationReport struct {
	CategoriesInserted int
	CategoriesSkipped  int
	DocumentsInserted  int
	DocumentsUpdated   int
	DocumentsSkipped   int
}

// BulkMigratePolicies copies `vatusa-old`.policy_categories and
// `vatusa-old`.policies into the Cobalt policy_category/policy_document
// tables. It is safe to re-run: categories are matched by title and
// documents by ident (their legacy natural keys -- neither Cobalt table has
// a legacy-id column), and it will not silently clobber a document a
// staffer has since edited through the new CRUD UI (see
// PolicyMigrationOptions.Force).
func BulkMigratePolicies(opts PolicyMigrationOptions) (PolicyMigrationReport, error) {
	ctx := context.Background()
	queries := dbconn.Queries()
	report := PolicyMigrationReport{}

	legacyCategories, err := queries.GetLegacyPolicyCategories(ctx)
	if err != nil {
		return report, fmt.Errorf("fetching legacy policy categories: %w", err)
	}
	legacyDocuments, err := queries.GetLegacyPolicies(ctx)
	if err != nil {
		return report, fmt.Errorf("fetching legacy policies: %w", err)
	}

	categoryIdMap, err := migrateCategories(ctx, queries, legacyCategories, opts, &report)
	if err != nil {
		return report, err
	}

	if err := migrateDocuments(ctx, queries, legacyDocuments, categoryIdMap, opts, &report); err != nil {
		return report, err
	}

	return report, nil
}

// migrateCategories inserts a policy_category for any legacy category whose
// title doesn't already exist. It never updates an existing category, with
// or without --force: policy_category carries no audit columns, so there is
// no way to tell "still exactly as the migration left it" apart from "a
// staffer already renamed/reordered it" the way migrateDocuments can via
// updated_by_cid. Matching by title also means renaming a migrated category
// before the migration is done running will produce a duplicate on the next
// run -- acceptable for a short-lived, pre-cutover migration window, but not
// a general-purpose sync.
func migrateCategories(
	ctx context.Context,
	queries *db.Queries,
	legacyCategories []db.GetLegacyPolicyCategoriesRow,
	opts PolicyMigrationOptions,
	report *PolicyMigrationReport,
) (map[uint32]int32, error) {
	categoryIdMap := make(map[uint32]int32, len(legacyCategories))

	for _, lc := range legacyCategories {
		title := strings.TrimSpace(lc.Name)

		existing, err := queries.GetPolicyCategoryByTitle(ctx, title)
		if err == nil {
			categoryIdMap[lc.ID] = existing.ID
			report.CategoriesSkipped++
			log.Printf("policy category %q already exists (id %d) -- skipping", title, existing.ID)
			continue
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("looking up policy category %q: %w", title, err)
		}

		if opts.DryRun {
			log.Printf("[dry-run] would create policy category %q (sort_order %d)", title, lc.Order)
			// No id to map to yet; downstream document dry-run logging
			// falls back to the legacy category id, which is fine since
			// nothing is actually written.
			categoryIdMap[lc.ID] = 0
			report.CategoriesInserted++
			continue
		}

		result, err := queries.CreatePolicyCategory(ctx, db.CreatePolicyCategoryParams{
			Title:     title,
			SortOrder: int32(lc.Order),
		})
		if err != nil {
			return nil, fmt.Errorf("creating policy category %q: %w", title, err)
		}
		newId, err := result.LastInsertId()
		if err != nil {
			return nil, fmt.Errorf("reading new id for policy category %q: %w", title, err)
		}
		categoryIdMap[lc.ID] = int32(newId)
		report.CategoriesInserted++
		log.Printf("created policy category %q (id %d)", title, newId)
	}

	return categoryIdMap, nil
}

// mappedPolicyDocument is the pure transform of one legacy `policies` row
// into the fields a policy_document write needs, split out from
// migrateDocuments so it can be unit tested without a database.
type mappedPolicyDocument struct {
	Ident         string
	Title         string
	Summary       string
	DocumentUrl   string
	Hidden        bool
	EffectiveDate time.Time
}

// mapLegacyDocument applies the legacy->Cobalt field mapping. hidden = NOT
// (visible = 1 AND perms = '0') -- legacy's canViewPolicy already gates
// every visible=0 row to VATUSA staff only, and the handful of
// visible=1-but-role-scoped rows (nonzero perms) have no equivalent in
// Cobalt's two-tier model, so per the product decision those also migrate
// hidden rather than public. fallbackDate is used when the legacy row's
// effective_date is NULL (schema allows it; no row does today).
func mapLegacyDocument(ld db.GetLegacyPoliciesRow, fallbackDate time.Time) mappedPolicyDocument {
	effectiveDate := fallbackDate
	if ld.EffectiveDate.Valid {
		effectiveDate = ld.EffectiveDate.Time
	}
	return mappedPolicyDocument{
		Ident:         strings.TrimSpace(ld.Ident),
		Title:         strings.TrimSpace(ld.Title),
		Summary:       html.UnescapeString(strings.TrimSpace(ld.Description)),
		DocumentUrl:   fmt.Sprintf("%s/docs/%s.%s", legacyDocsBaseURL, ld.Slug, ld.Extension),
		Hidden:        !(ld.Visible && ld.Perms == "0"),
		EffectiveDate: effectiveDate,
	}
}

// migrateDocuments inserts or refreshes a policy_document per legacy policy,
// keyed on ident (see mapLegacyDocument for the field-mapping rules).
func migrateDocuments(
	ctx context.Context,
	queries *db.Queries,
	legacyDocuments []db.GetLegacyPoliciesRow,
	categoryIdMap map[uint32]int32,
	opts PolicyMigrationOptions,
	report *PolicyMigrationReport,
) error {
	now := time.Now()

	for _, ld := range legacyDocuments {
		mapped := mapLegacyDocument(ld, now)
		ident, title, summary, documentUrl, hidden, effectiveDate :=
			mapped.Ident, mapped.Title, mapped.Summary, mapped.DocumentUrl, mapped.Hidden, mapped.EffectiveDate

		if !ld.EffectiveDate.Valid {
			log.Printf("policy %q (legacy id %d) has no effective_date -- defaulting to today", ident, ld.ID)
		}

		categoryId, ok := categoryIdMap[ld.Category]
		if !ok {
			return fmt.Errorf("policy %q (legacy id %d) references unknown legacy category %d", ident, ld.ID, ld.Category)
		}

		existing, err := queries.GetPolicyDocumentByIdent(ctx, ident)
		switch {
		case err == nil:
			if existing.UpdatedByCid != migrationActorCid && !opts.Force {
				report.DocumentsSkipped++
				log.Printf("policy document %q (id %d) was edited by cid %d -- skipping (use --force to overwrite)",
					ident, existing.ID, existing.UpdatedByCid)
				continue
			}
			if opts.DryRun {
				log.Printf("[dry-run] would update policy document %q (id %d)", ident, existing.ID)
				report.DocumentsUpdated++
				continue
			}
			result, err := queries.UpdatePolicyDocument(ctx, db.UpdatePolicyDocumentParams{
				PolicyCategoryID: categoryId,
				Ident:            ident,
				Title:            title,
				Summary:          summary,
				DocumentUrl:      documentUrl,
				EffectiveDate:    effectiveDate,
				Hidden:           hidden,
				SortOrder:        int32(ld.Order),
				UpdatedByCid:     migrationActorCid,
				UpdatedAt:        now,
				ID:               existing.ID,
			})
			if err != nil {
				return fmt.Errorf("updating policy document %q: %w", ident, err)
			}
			if rows, _ := result.RowsAffected(); rows == 0 {
				return fmt.Errorf("updating policy document %q: no rows affected", ident)
			}
			report.DocumentsUpdated++
			log.Printf("updated policy document %q (id %d)", ident, existing.ID)

		case errors.Is(err, sql.ErrNoRows):
			if opts.DryRun {
				log.Printf("[dry-run] would create policy document %q (hidden=%v)", ident, hidden)
				report.DocumentsInserted++
				continue
			}
			result, err := queries.CreatePolicyDocument(ctx, db.CreatePolicyDocumentParams{
				PolicyCategoryID: categoryId,
				Ident:            ident,
				Title:            title,
				Summary:          summary,
				DocumentUrl:      documentUrl,
				EffectiveDate:    effectiveDate,
				Hidden:           hidden,
				SortOrder:        int32(ld.Order),
				CreatedByCid:     migrationActorCid,
				UpdatedByCid:     migrationActorCid,
				CreatedAt:        now,
				UpdatedAt:        now,
			})
			if err != nil {
				return fmt.Errorf("creating policy document %q: %w", ident, err)
			}
			newId, err := result.LastInsertId()
			if err != nil {
				return fmt.Errorf("reading new id for policy document %q: %w", ident, err)
			}
			report.DocumentsInserted++
			log.Printf("created policy document %q (id %d, hidden=%v)", ident, newId, hidden)

		default:
			return fmt.Errorf("looking up policy document %q: %w", ident, err)
		}
	}

	return nil
}
