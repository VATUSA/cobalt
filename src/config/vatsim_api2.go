package config

import "os"

func VatsimApi2URL() string {
	return os.Getenv("VATSIM_API2_URL")
}

func VatsimApi2Key() string {
	return os.Getenv("VATSIM_API2_KEY")
}

func VatsimApi2Identifier() string {
	return os.Getenv("VATSIM_API2_IDENTIFIER")
}
