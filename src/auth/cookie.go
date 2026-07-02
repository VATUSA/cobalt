package auth

import (
	"net/http"
	"time"
	"vatusa-cobalt/config"
)

// NewSessionCookie builds the JWT session cookie with the correct domain for the environment.
func NewSessionCookie(value string) *http.Cookie {
	return &http.Cookie{
		Name:     JWTCookieName,
		Value:    value,
		Path:     "/",
		Domain:   config.CookieDomain(),
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(time.Hour * 24 * 7),
	}
}

// ClearSessionCookie builds an expired cookie that removes the JWT session cookie.
func ClearSessionCookie() *http.Cookie {
	return &http.Cookie{
		Name:     JWTCookieName,
		Value:    "",
		Path:     "/",
		Domain:   config.CookieDomain(),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	}
}
