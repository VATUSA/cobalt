package config

import "os"

func JWTKey() []byte {
	return []byte(os.Getenv("JWT_KEY"))
}

func CookieDomain() string {
	if val, ok := os.LookupEnv("COOKIE_DOMAIN"); ok {
		return val
	}
	if IsDevelopment() {
		return "localhost"
	}
	return "vatusa.net"
}
