package config

import (
	"net/url"
	"os"
	"strings"
)

func IsDevelopment() bool {
	appEnv := os.Getenv("APP_ENV")
	return appEnv == "dev"
}

func IsStaging() bool {
	appEnv := os.Getenv("APP_ENV")
	return appEnv == "staging"
}

func IsProduction() bool {
	appEnv := os.Getenv("APP_ENV")
	return appEnv == "prod" || appEnv == ""
}

func BaseURL() string {
	return os.Getenv("APP_BASE_URL")
}

func PostLoginURL() string {
	val, ok := os.LookupEnv("POST_LOGIN_URL")
	if !ok {
		return "https://vatusa.net"
	}
	return val
}

func StagingInternalURL() string {
	val, ok := os.LookupEnv("STAGING_INTERNAL_URL")
	if !ok {
		return "https://vatusa.dev/cobalt"
	}
	return val
}

// StagingPublicURL is the browser-reachable base URL of the staging/dev
// cobalt instance. Distinct from StagingInternalURL, which may point at an
// in-cluster service address for the prod->dev server-to-server /token/:cid
// call and is never reachable from the user's browser.
func StagingPublicURL() string {
	val, ok := os.LookupEnv("STAGING_PUBLIC_URL")
	if !ok {
		return "https://vatusa.dev/cobalt"
	}
	return val
}

func StagingActorToken() string {
	val, ok := os.LookupEnv("STAGING_ACTOR_TOKEN")
	if !ok {
		return ""
	}
	return val
}

// ConnectBaseURLOverride, when set, replaces the hardcoded VATSIM Connect base
// URL used for cobalt's own server-to-server calls (token exchange, user
// fetch). Used to point local/integration testing at a mock IdP.
func ConnectBaseURLOverride() string {
	return os.Getenv("VATSIM_CONNECT_BASE_URL")
}

// ConnectAuthorizeBaseURLOverride, when set, replaces the base URL used only
// for the browser-facing /oauth/authorize redirect. Distinct from
// ConnectBaseURLOverride because a test double reachable from inside the
// cobalt container (e.g. a Docker Compose service name) is not necessarily
// reachable by the browser following the redirect, and vice versa.
func ConnectAuthorizeBaseURLOverride() string {
	return os.Getenv("VATSIM_CONNECT_AUTHORIZE_BASE_URL")
}

// ConnectRedirectURIOverride, when set, replaces the hardcoded VATSIM Connect
// OAuth callback URL (the redirect_uri cobalt registers with the IdP).
func ConnectRedirectURIOverride() string {
	return os.Getenv("VATSIM_CONNECT_REDIRECT_URI")
}

// RelaysLoginToProd reports whether this instance must hand its logins to the
// production instance's staging relay (GetLoginForStaging) instead of running
// the VATSIM Connect round trip itself.
//
// This used to be plain IsStaging(), on the assumption that VATSIM had a
// redirect_uri registered only for cobalt.vatusa.net and therefore no other
// deployment could ever complete an OAuth round trip. That assumption has two
// costs which came due during the Azure migration:
//
//   - A dev environment cannot log anybody in unless production is healthy and
//     pointed at that exact instance, so dev inherits a prod dependency for the
//     one thing dev exists to test.
//   - The relay's target (STAGING_PUBLIC_URL / STAGING_INTERNAL_URL) is
//     single-valued and the internal hop uses in-cluster DNS, so only one dev
//     environment can have logins at a time and it has to share a cluster with
//     production. A second dev cluster cannot be reached at all.
//
// An organisation whose VATSIM Connect registration is approved can create
// additional OAuth clients itself, so a dev deployment can simply have its own.
// One that does sets VATSIM_CONNECT_REDIRECT_URI to its own callback and no
// longer needs the relay. Anything that sets nothing keeps the old behaviour
// byte for byte, which is what keeps the existing DOKS dev instance working.
//
// Note there is deliberately no separate on/off flag: the override *is* the
// thing that makes a direct round trip possible, so a second variable could
// only ever disagree with it. If the override is set without a matching
// VATSIM_CONNECT_CLIENT_ID/SECRET for the same client, the round trip fails at
// VATSIM with a redirect_uri mismatch — which is a clearer signal than silently
// falling back to a relay that the operator did not ask for.
func RelaysLoginToProd() bool {
	return IsStaging() && ConnectRedirectURIOverride() == ""
}

func RedirectAllowlist() []string {
	val, ok := os.LookupEnv("REDIRECT_ALLOWLIST")
	if !ok || val == "" {
		return nil
	}
	hosts := strings.Split(val, ",")
	for i, h := range hosts {
		hosts[i] = strings.TrimSpace(h)
	}
	return hosts
}

// IsAllowedRedirect reports whether target is a parseable https URL (http
// permitted only in development) whose host matches or is a subdomain of an
// entry in RedirectAllowlist. Used to validate caller-supplied post-login
// redirect targets before honoring them, to prevent open redirects.
func IsAllowedRedirect(target string) bool {
	if target == "" {
		return false
	}
	u, err := url.Parse(target)
	if err != nil || u.Hostname() == "" {
		return false
	}
	if u.Scheme != "https" && !(IsDevelopment() && u.Scheme == "http") {
		return false
	}
	host := u.Hostname()
	for _, allowed := range RedirectAllowlist() {
		if allowed == "" {
			continue
		}
		if host == allowed || strings.HasSuffix(host, "."+allowed) {
			return true
		}
	}
	return false
}

// IsSafeDocumentURL reports whether target is a parseable https URL (http
// permitted only in development), with no host allowlist. Used to validate a
// caller-supplied policy document_url before storing it, since it is later
// rendered as an href — without this check a scheme like javascript: would be
// stored and executed in the staff app.
func IsSafeDocumentURL(target string) bool {
	if target == "" {
		return false
	}
	u, err := url.Parse(target)
	if err != nil || u.Hostname() == "" {
		return false
	}
	return u.Scheme == "https" || (IsDevelopment() && u.Scheme == "http")
}
