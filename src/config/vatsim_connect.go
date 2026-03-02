package config

import "os"

func ConnectClientId() string {
	return os.Getenv("VATSIM_CONNECT_CLIENT_ID")
}

func ConnectClientSecret() string {
	return os.Getenv("VATSIM_CONNECT_CLIENT_SECRET")
}

const ConnectTimestampFormat = "2006-01-02T15:04:05"
