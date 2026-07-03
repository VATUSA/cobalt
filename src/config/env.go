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

func StagingActorToken() string {
	val, ok := os.LookupEnv("STAGING_ACTOR_TOKEN")
	if !ok {
		return ""
	}
	return val
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
