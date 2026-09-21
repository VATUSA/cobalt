package config

import (
	"os"
	"testing"
)

func TestConnectionStringTLS(t *testing.T) {
	for _, env := range []string{"DB_HOST", "DB_PORT", "DB_USER", "DB_PASS", "DB_NAME", "DB_TLS"} {
		t.Setenv(env, "")
	}
	t.Setenv("DB_HOST", "db.example.com")
	t.Setenv("DB_PORT", "3306")
	t.Setenv("DB_USER", "u")
	t.Setenv("DB_PASS", "p")
	t.Setenv("DB_NAME", "n")

	tests := []struct {
		name  string
		dbTLS string
		want  string
	}{
		// Unset must reproduce the pre-TLS DSN exactly: DigitalOcean deployments,
		// production included, set no DB_TLS and must not start negotiating TLS.
		{"unset omits the parameter", "", "u:p@tcp(db.example.com:3306)/n?parseTime=true"},
		{"true verifies fully", "true", "u:p@tcp(db.example.com:3306)/n?parseTime=true&tls=true"},
		{"skip-verify encrypts only", "skip-verify", "u:p@tcp(db.example.com:3306)/n?parseTime=true&tls=skip-verify"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.dbTLS == "" {
				os.Unsetenv("DB_TLS")
			} else {
				t.Setenv("DB_TLS", tt.dbTLS)
			}
			if got := ConnectionString(); got != tt.want {
				t.Errorf("ConnectionString() = %q, want %q", got, tt.want)
			}
		})
	}
}
