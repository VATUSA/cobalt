package legacy_migration

import (
	"database/sql"
	"testing"
	"time"
	"vatusa-cobalt/db"
)

func TestMapLegacyDocument_Hidden(t *testing.T) {
	cases := []struct {
		name    string
		visible bool
		perms   string
		want    bool
	}{
		{"visible with no perms restriction is public", true, "0", false},
		{"visible with a single role restriction is hidden", true, "1", true},
		{"visible with a pipe-delimited role restriction is hidden", true, "2|3|4|6", true},
		{"not visible is hidden regardless of perms", false, "0", true},
		{"not visible with a role restriction is hidden", false, "5", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ld := db.GetLegacyPoliciesRow{
				Visible: tc.visible,
				Perms:   tc.perms,
			}
			got := mapLegacyDocument(ld, time.Now())
			if got.Hidden != tc.want {
				t.Errorf("Hidden = %v, want %v (visible=%v perms=%q)", got.Hidden, tc.want, tc.visible, tc.perms)
			}
		})
	}
}

func TestMapLegacyDocument_DocumentURL(t *testing.T) {
	ld := db.GetLegacyPoliciesRow{
		Slug:      "vatusa-organizational-chart",
		Extension: "jpeg",
	}
	got := mapLegacyDocument(ld, time.Now())
	want := "https://vatusa-storage.nyc3.cdn.digitaloceanspaces.com/docs/vatusa-organizational-chart.jpeg"
	if got.DocumentUrl != want {
		t.Errorf("DocumentUrl = %q, want %q", got.DocumentUrl, want)
	}
}

func TestMapLegacyDocument_TrimsIdentAndTitle(t *testing.T) {
	ld := db.GetLegacyPoliciesRow{
		Ident: "  BSG2021.1 ",
		Title: " VATUSA Brand Style Guide\n",
	}
	got := mapLegacyDocument(ld, time.Now())
	if got.Ident != "BSG2021.1" {
		t.Errorf("Ident = %q, want trimmed %q", got.Ident, "BSG2021.1")
	}
	if got.Title != "VATUSA Brand Style Guide" {
		t.Errorf("Title = %q, want trimmed %q", got.Title, "VATUSA Brand Style Guide")
	}
}

func TestMapLegacyDocument_SummaryDecodesEntitiesAndTrims(t *testing.T) {
	ld := db.GetLegacyPoliciesRow{
		Description: "  VATUSA Branding &amp; Styling Guidelines  ",
	}
	got := mapLegacyDocument(ld, time.Now())
	want := "VATUSA Branding & Styling Guidelines"
	if got.Summary != want {
		t.Errorf("Summary = %q, want %q", got.Summary, want)
	}
}

func TestMapLegacyDocument_SummaryLeavesPlainAmpersandAlone(t *testing.T) {
	// A raw "&" that isn't part of a recognized entity sequence must survive
	// unescaping unchanged -- html.UnescapeString only touches actual
	// entities, and most legacy descriptions contain a literal "&" rather
	// than an encoded one.
	ld := db.GetLegacyPoliciesRow{
		Description: "VATUSA Branding & Styling Guidelines v.2021.1",
	}
	got := mapLegacyDocument(ld, time.Now())
	want := "VATUSA Branding & Styling Guidelines v.2021.1"
	if got.Summary != want {
		t.Errorf("Summary = %q, want %q", got.Summary, want)
	}
}

func TestMapLegacyDocument_EffectiveDateFromLegacy(t *testing.T) {
	legacyDate := time.Date(2021, 6, 21, 0, 0, 0, 0, time.UTC)
	ld := db.GetLegacyPoliciesRow{
		EffectiveDate: sql.NullTime{Time: legacyDate, Valid: true},
	}
	fallback := time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)
	got := mapLegacyDocument(ld, fallback)
	if !got.EffectiveDate.Equal(legacyDate) {
		t.Errorf("EffectiveDate = %v, want %v", got.EffectiveDate, legacyDate)
	}
}

func TestMapLegacyDocument_EffectiveDateFallsBackWhenNull(t *testing.T) {
	fallback := time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)
	ld := db.GetLegacyPoliciesRow{
		EffectiveDate: sql.NullTime{Valid: false},
	}
	got := mapLegacyDocument(ld, fallback)
	if !got.EffectiveDate.Equal(fallback) {
		t.Errorf("EffectiveDate = %v, want fallback %v", got.EffectiveDate, fallback)
	}
}
