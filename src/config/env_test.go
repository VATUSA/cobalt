package config

import "testing"

// TestRelaysLoginToProd pins the decision that used to be a bare IsStaging()
// check. The case that matters most is the first one: a staging deployment that
// sets no override must still relay, because that is the existing DOKS dev
// instance and this change must not alter it.
func TestRelaysLoginToProd(t *testing.T) {
	cases := []struct {
		name        string
		appEnv      string
		redirectURI string
		want        bool
	}{
		{
			name:        "staging with no own client relays, as before",
			appEnv:      "staging",
			redirectURI: "",
			want:        true,
		},
		{
			name:        "staging with its own client goes direct",
			appEnv:      "staging",
			redirectURI: "https://cobalt.azure.vatusa.dev/login/connect",
			want:        false,
		},
		{
			name:        "an override set to empty is the same as unset",
			appEnv:      "staging",
			redirectURI: "",
			want:        true,
		},
		{
			name:        "prod never relays to itself",
			appEnv:      "prod",
			redirectURI: "",
			want:        false,
		},
		{
			name:        "prod with an override still never relays",
			appEnv:      "prod",
			redirectURI: "https://cobalt.vatusa.net/login/connect",
			want:        false,
		},
		{
			name:        "unset APP_ENV means prod, so no relay",
			appEnv:      "",
			redirectURI: "",
			want:        false,
		},
		{
			name:        "local development does not relay",
			appEnv:      "dev",
			redirectURI: "",
			want:        false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("APP_ENV", tc.appEnv)
			t.Setenv("VATSIM_CONNECT_REDIRECT_URI", tc.redirectURI)
			if got := RelaysLoginToProd(); got != tc.want {
				t.Errorf("RelaysLoginToProd() = %v, want %v (APP_ENV=%q, VATSIM_CONNECT_REDIRECT_URI=%q)",
					got, tc.want, tc.appEnv, tc.redirectURI)
			}
		})
	}
}
