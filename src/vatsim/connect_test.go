package vatsim

import "testing"

// TestConnectRedirectURI guards the callback URL that VATSIM matches against
// the registered client. RelaysLoginToProd now keys off the override, so a
// regression here would take a deployment's login with it: an instance would
// stop relaying and then present a redirect_uri VATSIM rejects.
func TestConnectRedirectURI(t *testing.T) {
	cases := []struct {
		name     string
		appEnv   string
		override string
		want     string
	}{
		{
			name:   "prod uses the registered cobalt.vatusa.net callback",
			appEnv: "prod",
			want:   "https://cobalt.vatusa.net/login/connect",
		},
		{
			name:   "staging without an override falls back to prod's callback",
			appEnv: "staging",
			want:   "https://cobalt.vatusa.net/login/connect",
		},
		{
			name:   "local development points at localhost",
			appEnv: "dev",
			want:   "http://localhost:8000/cobalt/login/connect",
		},
		{
			name:     "an override wins on a staging instance with its own client",
			appEnv:   "staging",
			override: "https://cobalt.azure.vatusa.dev/login/connect",
			want:     "https://cobalt.azure.vatusa.dev/login/connect",
		},
		{
			name:     "an override wins over the localhost default too",
			appEnv:   "dev",
			override: "http://127.0.0.1:9000/login/connect",
			want:     "http://127.0.0.1:9000/login/connect",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("APP_ENV", tc.appEnv)
			t.Setenv("VATSIM_CONNECT_REDIRECT_URI", tc.override)
			if got := ConnectRedirectURI(); got != tc.want {
				t.Errorf("ConnectRedirectURI() = %q, want %q", got, tc.want)
			}
		})
	}
}
