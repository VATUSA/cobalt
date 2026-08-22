package auth

import "testing"

// TestCreateAndGetCIDFromToken covers the session token round trip: a token
// minted by CreateToken must decode back to the same cid via
// GetCIDFromToken.
func TestCreateAndGetCIDFromToken(t *testing.T) {
	t.Setenv("JWT_KEY", "test-signing-key")

	token, err := CreateToken(123456, "Test User", "event:write", "ZDV:event:write", true, "ZDV")
	if err != nil {
		t.Fatalf("CreateToken() error = %v", err)
	}

	cid, err := GetCIDFromToken(token)
	if err != nil {
		t.Fatalf("GetCIDFromToken() error = %v", err)
	}
	if cid != 123456 {
		t.Errorf("GetCIDFromToken() = %d, want 123456", cid)
	}
}

func TestGetCIDFromToken_WrongKeyRejected(t *testing.T) {
	t.Setenv("JWT_KEY", "key-a")
	token, err := CreateToken(1, "", "", "", false, "")
	if err != nil {
		t.Fatalf("CreateToken() error = %v", err)
	}

	t.Setenv("JWT_KEY", "key-b")
	if _, err := GetCIDFromToken(token); err == nil {
		t.Error("expected an error decoding a token signed with a different key")
	}
}

func TestGetCIDFromToken_GarbageTokenRejected(t *testing.T) {
	t.Setenv("JWT_KEY", "test-signing-key")
	if _, err := GetCIDFromToken("not.a.jwt"); err == nil {
		t.Error("expected an error decoding a malformed token")
	}
}

func TestGetCIDFromToken_RejectsAlgNone(t *testing.T) {
	// jwt.WithValidMethods should reject a token asserting "alg: none" even
	// though its signature would trivially "validate" - a classic JWT
	// downgrade attack.
	t.Setenv("JWT_KEY", "test-signing-key")
	// header {"alg":"none","typ":"JWT"}, payload {"cid":999}, no signature
	noneToken := "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJjaWQiOjk5OX0."
	if _, err := GetCIDFromToken(noneToken); err == nil {
		t.Error("expected alg=none tokens to be rejected")
	}
}
