package config

import "testing"

// TestIsAllowedRedirect covers the open-redirect guard that gates every
// caller-supplied `redirect` query param across the login flow. A mistake
// here (e.g. matching by substring instead of host suffix) turns the login
// flow into an open redirect.
func TestIsAllowedRedirect(t *testing.T) {
	t.Setenv("REDIRECT_ALLOWLIST", "vatusa.net,vatusa.dev")

	cases := []struct {
		name   string
		target string
		want   bool
	}{
		{"exact host match", "https://vatusa.net/callback", true},
		{"subdomain of allowed host", "https://www.vatusa.net/callback", true},
		{"nested subdomain of allowed host", "https://staff.portal.vatusa.dev/x", true},
		{"different allowed entry", "https://vatusa.dev/x", true},
		{"unrelated host", "https://evil.com/callback", false},
		{"host that merely ends with allowed suffix, no dot boundary", "https://notvatusa.net/callback", false},
		{"host containing allowed string as substring, not a suffix", "https://vatusa.net.evil.com/callback", false},
		{"http scheme rejected outside development", "http://vatusa.net/callback", false},
		{"empty target", "", false},
		{"unparseable target", "://bad-url", false},
		{"missing scheme entirely", "vatusa.net/callback", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsAllowedRedirect(tc.target); got != tc.want {
				t.Errorf("IsAllowedRedirect(%q) = %v, want %v", tc.target, got, tc.want)
			}
		})
	}
}

func TestIsAllowedRedirect_HTTPPermittedInDevelopment(t *testing.T) {
	t.Setenv("REDIRECT_ALLOWLIST", "vatusa.net")
	t.Setenv("APP_ENV", "dev")

	if !IsAllowedRedirect("http://vatusa.net/callback") {
		t.Error("expected http to be permitted for an allowed host in development")
	}
	if IsAllowedRedirect("http://evil.com/callback") {
		t.Error("did not expect an unrelated host to be allowed even in development")
	}
}

func TestIsAllowedRedirect_EmptyAllowlistDeniesEverything(t *testing.T) {
	t.Setenv("REDIRECT_ALLOWLIST", "")
	if IsAllowedRedirect("https://vatusa.net/callback") {
		t.Error("expected every target to be denied with an empty allowlist")
	}
}

func TestRedirectAllowlist_TrimsWhitespace(t *testing.T) {
	t.Setenv("REDIRECT_ALLOWLIST", " vatusa.net , vatusa.dev ")
	got := RedirectAllowlist()
	want := []string{"vatusa.net", "vatusa.dev"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("entry %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestRedirectAllowlist_UnsetReturnsNil(t *testing.T) {
	t.Setenv("REDIRECT_ALLOWLIST", "")
	if got := RedirectAllowlist(); got != nil {
		t.Errorf("expected nil allowlist when unset, got %v", got)
	}
}

func TestAppEnvHelpers(t *testing.T) {
	cases := []struct {
		env            string
		wantDev        bool
		wantStaging    bool
		wantProduction bool
	}{
		{"dev", true, false, false},
		{"staging", false, true, false},
		{"prod", false, false, true},
		{"", false, false, true}, // empty APP_ENV defaults to production
		{"something_else", false, false, false},
	}

	for _, tc := range cases {
		t.Run("APP_ENV="+tc.env, func(t *testing.T) {
			t.Setenv("APP_ENV", tc.env)
			if got := IsDevelopment(); got != tc.wantDev {
				t.Errorf("IsDevelopment() = %v, want %v", got, tc.wantDev)
			}
			if got := IsStaging(); got != tc.wantStaging {
				t.Errorf("IsStaging() = %v, want %v", got, tc.wantStaging)
			}
			if got := IsProduction(); got != tc.wantProduction {
				t.Errorf("IsProduction() = %v, want %v", got, tc.wantProduction)
			}
		})
	}
}
