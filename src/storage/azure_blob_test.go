package storage

import (
	"net/http"
	"strings"
	"testing"
)

// End-to-end vector for Azure Blob's Shared Key authorization scheme,
// independently computed via openssl against Azurite's well-known emulator
// account key (never used against real Azure — picked only because it's a
// public, well-known key, so this test has no real secret in it). This is
// the check that actually proves the string-to-sign and signature are
// assembled correctly — mirrors sigv4_test.go's approach for Spaces.
func TestSignAzureBlobRequestExample(t *testing.T) {
	const account = "devstoreaccount1"
	const accountKey = "Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw=="
	const blobPath = "vatusa-events/event-banners/zdv/deadbeefdeadbeefdeadbeefdeadbeef.png"

	req, err := http.NewRequest(http.MethodPut, "https://"+account+".blob.core.windows.net/"+blobPath, strings.NewReader("test"))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "image/png")
	req.Header.Set("x-ms-blob-type", "BlockBlob")
	req.Header.Set("x-ms-date", "Wed, 09 Sep 2026 12:00:00 GMT")
	req.Header.Set("x-ms-version", "2021-08-06")

	if err := signAzureBlobRequest(req, account, accountKey, blobPath, 4); err != nil {
		t.Fatal(err)
	}

	want := "SharedKey devstoreaccount1:QG91JR0xNMn4RSyFpSJSSmWYJ7VE+vrEKE2MXmrmbZE="
	if got := req.Header.Get("Authorization"); got != want {
		t.Errorf("Authorization = %s, want %s", got, want)
	}
}

// A malformed (non-base64) account key must be reported, not silently
// produce a garbage signature that Azure would just reject with a 403.
func TestSignAzureBlobRequestBadKey(t *testing.T) {
	req, err := http.NewRequest(http.MethodPut, "https://account.blob.core.windows.net/container/key", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := signAzureBlobRequest(req, "account", "not-valid-base64!!", "container/key", 0); err == nil {
		t.Fatal("expected an error for a non-base64 account key")
	}
}
