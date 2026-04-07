package config

import "os"

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
		return "https://cobalt-service.cobalt-dev.svc.cluster.local:8080"
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
