package auth

import "testing"

// A zero-value cache (tokenActorMap still nil, as it is before the
// background loader's first Load completes) must be safe to read from —
// this used to panic on a nil *map dereference while holding the read lock,
// which then deadlocked every future Load/Get call permanently.
func TestTokenActorCacheGetBeforeLoad(t *testing.T) {
	var cache TokenActorCache

	actor, ok := cache.Get("some-token")
	if ok {
		t.Errorf("got ok=true, want false for an empty cache")
	}
	if actor != nil {
		t.Errorf("got actor=%v, want nil for an empty cache", actor)
	}
}
