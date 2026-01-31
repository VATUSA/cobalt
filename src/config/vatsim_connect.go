package config

import "os"

func ConnectClientId() string {
	return os.Getenv("VATSIM_CONNECT_CLIENT_ID")
}

func ConnectClientSecret() string {
	return os.Getenv("VATSIM_CONNECT_CLIENT_SECRET")
}
