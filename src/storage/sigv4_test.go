package storage

import (
	"encoding/hex"
	"net/http"
	"strings"
	"testing"
	"time"
)

// Signing key derivation vector from the AWS "deriving a signing key"
// documentation.
func TestDeriveSigningKey(t *testing.T) {
	key := deriveSigningKey("wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY", "20120215", "us-east-1", "iam")
	got := hex.EncodeToString(key)
	want := "f4780e2d9f65fa895f9c67b32ce1baf0b0d8a43505a000a1a9e090d414db404d"
	if got != want {
		t.Fatalf("signing key = %s, want %s", got, want)
	}
}

// End-to-end vector from AWS's documented "PUT Object" SigV4 example. This is
// the check that actually proves the canonical request, string to sign and
// signature are all assembled correctly.
func TestSignRequestPutObjectExample(t *testing.T) {
	body := []byte("Welcome to Amazon S3.")

	req, err := http.NewRequest(http.MethodPut, "https://examplebucket.s3.amazonaws.com/test$file.text", strings.NewReader(string(body)))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Date", "Fri, 24 May 2013 00:00:00 GMT")
	req.Header.Set("X-Amz-Storage-Class", "REDUCED_REDUNDANCY")

	signedAt := time.Date(2013, 5, 24, 0, 0, 0, 0, time.UTC)
	signRequest(req, "AKIAIOSFODNN7EXAMPLE", "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY", "us-east-1", "s3", hashHex(body), signedAt)

	wantHash := "44ce7dd67c959e0d3524ffac1771dfbba87d2b6b4b4e99e42034a8b803f8b072"
	if got := req.Header.Get("X-Amz-Content-Sha256"); got != wantHash {
		t.Errorf("payload hash = %s, want %s", got, wantHash)
	}

	// The canonical request AWS documents for this example, verbatim. If this
	// matches, the signature below can only differ by the key derivation,
	// which TestDeriveSigningKey pins separately.
	wantCanonical := strings.Join([]string{
		"PUT",
		"/test%24file.text",
		"",
		"date:Fri, 24 May 2013 00:00:00 GMT",
		"host:examplebucket.s3.amazonaws.com",
		"x-amz-content-sha256:" + wantHash,
		"x-amz-date:20130524T000000Z",
		"x-amz-storage-class:REDUCED_REDUNDANCY",
		"",
		"date;host;x-amz-content-sha256;x-amz-date;x-amz-storage-class",
		wantHash,
	}, "\n")
	_, gotCanonical := canonicalRequest(req, wantHash)
	if gotCanonical != wantCanonical {
		t.Errorf("canonical request =\n%q\nwant\n%q", gotCanonical, wantCanonical)
	}
	if got := hashHex([]byte(gotCanonical)); got != "9e0e90d9c76de8fa5b200d8c849cd5b8dc7a3be3951ddb7f6a76b4158342019d" {
		t.Errorf("canonical request hash = %s, want AWS documented value", got)
	}

	want := "AWS4-HMAC-SHA256 " +
		"Credential=AKIAIOSFODNN7EXAMPLE/20130524/us-east-1/s3/aws4_request, " +
		"SignedHeaders=date;host;x-amz-content-sha256;x-amz-date;x-amz-storage-class, " +
		"Signature=7c0f3caf24a16d5948905b8ebf67d29fb415e93fddaed9ca6aeb5ac2348cfee4"
	if got := req.Header.Get("Authorization"); got != want {
		t.Errorf("Authorization =\n%s\nwant\n%s", got, want)
	}
}

func TestCanonicalURI(t *testing.T) {
	cases := map[string]string{
		"":                          "/",
		"/":                         "/",
		"/test$file.text":           "/test%24file.text",
		"/event-banners/zdv/ab.png": "/event-banners/zdv/ab.png",
		"/a b":                      "/a%20b",
		"/a%20b":                    "/a%20b",
	}
	for in, want := range cases {
		if got := canonicalURI(in); got != want {
			t.Errorf("canonicalURI(%q) = %q, want %q", in, got, want)
		}
	}
}
