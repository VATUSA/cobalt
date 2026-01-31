package config

import "os"

func IsDevelopment() bool {
	appEnv := os.Getenv("APP_ENV")
	return appEnv == "dev"
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
