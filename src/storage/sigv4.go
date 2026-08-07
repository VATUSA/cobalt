package storage

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
)

// Minimal AWS Signature Version 4 signing, enough for the single PutObject
// call we make against DigitalOcean Spaces. The full AWS SDK would work too,
// but it drags in ~15 modules and its default flexible-checksum behaviour has
// to be disabled for non-AWS S3 endpoints — for one request type it's less
// moving parts to sign it ourselves.

const (
	sigV4Algorithm = "AWS4-HMAC-SHA256"
	isoDateTime    = "20060102T150405Z"
	isoDate        = "20060102"
)

// signRequest adds the x-amz-date, x-amz-content-sha256 and Authorization
// headers to req. payloadHash must be the lowercase hex SHA-256 of the exact
// bytes that will be sent as the body.
func signRequest(req *http.Request, accessKey, secretKey, region, service, payloadHash string, now time.Time) {
	now = now.UTC()
	amzDate := now.Format(isoDateTime)
	dateStamp := now.Format(isoDate)

	req.Header.Set("X-Amz-Date", amzDate)
	req.Header.Set("X-Amz-Content-Sha256", payloadHash)
	if req.Host != "" {
		req.Header.Set("Host", req.Host)
	} else {
		req.Header.Set("Host", req.URL.Host)
	}

	signedHeaders, canonicalRequest := canonicalRequest(req, payloadHash)

	credentialScope := strings.Join([]string{dateStamp, region, service, "aws4_request"}, "/")
	stringToSign := strings.Join([]string{
		sigV4Algorithm,
		amzDate,
		credentialScope,
		hashHex([]byte(canonicalRequest)),
	}, "\n")

	signingKey := deriveSigningKey(secretKey, dateStamp, region, service)
	signature := hex.EncodeToString(hmacSHA256(signingKey, stringToSign))

	req.Header.Set("Authorization", fmt.Sprintf(
		"%s Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		sigV4Algorithm, accessKey, credentialScope, signedHeaders, signature,
	))
}

// canonicalRequest returns the signed header list and the canonical request
// string that gets hashed into the string to sign.
func canonicalRequest(req *http.Request, payloadHash string) (signedHeaders, canonical string) {
	signedHeaders, canonicalHeaders := canonicalizeHeaders(req)
	return signedHeaders, strings.Join([]string{
		req.Method,
		canonicalURI(req.URL.EscapedPath()),
		req.URL.Query().Encode(),
		canonicalHeaders,
		signedHeaders,
		payloadHash,
	}, "\n")
}

// canonicalizeHeaders returns the semicolon-joined signed header list and the
// canonical header block, both derived from every header currently set on the
// request. Signing everything we send (rather than a fixed subset) keeps the
// signature valid if a caller adds a header later.
func canonicalizeHeaders(req *http.Request) (signedHeaders, canonicalHeaders string) {
	names := make([]string, 0, len(req.Header)+1)
	values := make(map[string]string, len(req.Header)+1)

	for name, vals := range req.Header {
		lower := strings.ToLower(name)
		if lower == "authorization" {
			continue
		}
		trimmed := make([]string, len(vals))
		for i, v := range vals {
			trimmed[i] = strings.Join(strings.Fields(v), " ")
		}
		names = append(names, lower)
		values[lower] = strings.Join(trimmed, ",")
	}

	sort.Strings(names)

	var block strings.Builder
	for _, name := range names {
		block.WriteString(name)
		block.WriteString(":")
		block.WriteString(values[name])
		block.WriteString("\n")
	}

	return strings.Join(names, ";"), block.String()
}

// canonicalURI re-encodes an already-escaped path the way SigV4 expects:
// RFC 3986 unreserved characters stay literal, everything else is
// percent-encoded uppercase, and path separators are preserved. Go's
// EscapedPath leaves some sub-delims (e.g. "!", "$", "&") unescaped, so we
// decode each segment and re-encode it strictly.
func canonicalURI(escapedPath string) string {
	if escapedPath == "" {
		return "/"
	}
	segments := strings.Split(escapedPath, "/")
	for i, seg := range segments {
		segments[i] = uriEncode(unescapeSegment(seg))
	}
	return strings.Join(segments, "/")
}

func unescapeSegment(seg string) string {
	var out strings.Builder
	for i := 0; i < len(seg); i++ {
		if seg[i] == '%' && i+2 < len(seg) {
			if b, err := hex.DecodeString(seg[i+1 : i+3]); err == nil {
				out.Write(b)
				i += 2
				continue
			}
		}
		out.WriteByte(seg[i])
	}
	return out.String()
}

func uriEncode(s string) string {
	var out strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= 'A' && c <= 'Z', c >= 'a' && c <= 'z', c >= '0' && c <= '9',
			c == '-', c == '_', c == '.', c == '~':
			out.WriteByte(c)
		default:
			fmt.Fprintf(&out, "%%%02X", c)
		}
	}
	return out.String()
}

func deriveSigningKey(secretKey, dateStamp, region, service string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+secretKey), dateStamp)
	kRegion := hmacSHA256(kDate, region)
	kService := hmacSHA256(kRegion, service)
	return hmacSHA256(kService, "aws4_request")
}

func hmacSHA256(key []byte, data string) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	return h.Sum(nil)
}

func hashHex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
