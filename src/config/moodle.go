package config

import "os"

// MoodleURL is the base URL of the VATUSA Academy Moodle instance, e.g.
// https://academy.vatusa.net. Empty disables Moodle sync entirely.
func MoodleURL() string {
	return os.Getenv("MOODLE_URL")
}

// MoodleToken is the Moodle web service token used for user/cohort
// administration calls (core_user_*, core_cohort_*).
func MoodleToken() string {
	return os.Getenv("MOODLE_TOKEN")
}

// MoodleEnabled reports whether Moodle sync should run at all: only on
// prod/staging (matching the legacy api behavior), and only once the
// instance's URL and token are configured.
func MoodleEnabled() bool {
	return (IsProduction() || IsStaging()) && MoodleURL() != "" && MoodleToken() != ""
}
