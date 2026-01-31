package config

import (
	"fmt"
	"os"
)

func ConnectionString() string {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASS")
	database := os.Getenv("DB_NAME")

	cs := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, password, host, port, database)
	return cs
}
