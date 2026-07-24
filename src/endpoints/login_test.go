package endpoints

import "testing"

// TestParseStagingRelayState covers the distinction Connect() depends on: a
// relay-tagged state must be recognized as a relay, and a bare caller
// redirect must not be. Conflating the two relayed prod logins through the
// dev instance's /token/:cid.
func TestParseStagingRelayState(t *testing.T) {
	cases := []struct {
		name         string
		state        string
		wantRedirect string
		wantIsRelay  bool
	}{
		{
			name:         "relay with no caller redirect",
			state:        stagingRelayState(""),
			wantRedirect: "",
			wantIsRelay:  true,
		},
		{
			name:         "relay carrying a caller redirect",
			state:        stagingRelayState("https://www.vatusa.dev/legacy/auth/callback"),
			wantRedirect: "https://www.vatusa.dev/legacy/auth/callback",
			wantIsRelay:  true,
		},
		{
			name:         "plain prod login with its own redirect is not a relay",
			state:        "https://www.vatusa.net/legacy/auth/callback",
			wantRedirect: "",
			wantIsRelay:  false,
		},
		{
			name:         "empty state is not a relay",
			state:        "",
			wantRedirect: "",
			wantIsRelay:  false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			redirect, isRelay := parseStagingRelayState(tc.state)
			if isRelay != tc.wantIsRelay {
				t.Errorf("isRelay = %v, want %v (state %q)", isRelay, tc.wantIsRelay, tc.state)
			}
			if redirect != tc.wantRedirect {
				t.Errorf("redirect = %q, want %q", redirect, tc.wantRedirect)
			}
		})
	}
}
