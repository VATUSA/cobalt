package config

import (
	"fmt"
	"os"
	"strings"
)

// DigitalOcean Spaces is S3-compatible object storage. We use it to host
// user-uploaded event banners so we're not depending on Imgur (blocked in
// several jurisdictions, including the UK) or Discord (which actively
// discourages being used as a CDN for off-app browsers).

func SpacesKey() string {
	return os.Getenv("DO_SPACES_KEY")
}

func SpacesSecret() string {
	return os.Getenv("DO_SPACES_SECRET")
}

// SpacesRegion defaults to sfo3, where the vatusa-events bucket lives. Note
// that most other VATUSA buckets are in nyc3 — this one is not.
func SpacesRegion() string {
	val, ok := os.LookupEnv("DO_SPACES_REGION")
	if !ok || val == "" {
		return "sfo3"
	}
	return val
}

func SpacesBucket() string {
	return os.Getenv("DO_SPACES_BUCKET")
}

// SpacesEndpoint is the virtual-hosted-style origin for the bucket, e.g.
// https://vatusa-events.sfo3.digitaloceanspaces.com. Derived from the
// bucket and region unless DO_SPACES_ENDPOINT overrides it.
func SpacesEndpoint() string {
	val, ok := os.LookupEnv("DO_SPACES_ENDPOINT")
	if ok && val != "" {
		return strings.TrimSuffix(val, "/")
	}
	return fmt.Sprintf("https://%s.%s.digitaloceanspaces.com", SpacesBucket(), SpacesRegion())
}

// SpacesPublicBaseURL is the base URL written into event records. It defaults
// to the bucket origin but should be set to the Spaces CDN endpoint (or a
// custom domain in front of it) in deployed environments so banners are served
// from the edge rather than the origin.
func SpacesPublicBaseURL() string {
	val, ok := os.LookupEnv("DO_SPACES_PUBLIC_BASE_URL")
	if ok && val != "" {
		return strings.TrimSuffix(val, "/")
	}
	return SpacesEndpoint()
}

// IsSpacesConfigured reports whether uploads can be served. When false the
// event endpoints fall back to accepting a caller-supplied banner URL only.
func IsSpacesConfigured() bool {
	return SpacesKey() != "" && SpacesSecret() != "" && SpacesBucket() != ""
}
