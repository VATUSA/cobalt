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
