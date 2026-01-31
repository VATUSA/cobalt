package config

import "os"

func JWTKey() []byte {
	return []byte(os.Getenv("JWT_KEY"))
}

func CookieDomain() string {
	if IsDevelopment() {
		return "localhost"
	}
	return "vatusa.net"
}
