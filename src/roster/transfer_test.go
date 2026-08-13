package roster

import (
	"strings"
	"testing"
)

func TestResolveActorDisplayName(t *testing.T) {
	cases := []struct {
		name     string
		actorCid int64
		wantName string
		wantErr  string
	}{
		{name: "system actor", actorCid: 0, wantName: "Automated"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			name, err := resolveActorDisplayName(tc.actorCid)
			if tc.wantErr == "" && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tc.wantErr != "" && err == nil {
				t.Errorf("expected error containing %q, got nil", tc.wantErr)
			}
			if tc.wantErr != "" && err != nil && !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("error = %q, want substring %q", err.Error(), tc.wantErr)
			}
			if name != tc.wantName {
				t.Errorf("name = %q, want %q", name, tc.wantName)
			}
		})
	}
}
