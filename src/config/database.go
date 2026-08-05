package config

import (
	"fmt"
	"os"
	"strconv"
)

// defaultMaxOpenConns bounds the connection pool. database/sql defaults to
// unlimited, which is the wrong default here: the MySQL server is shared with
// every other VATUSA service, so an unbounded pool lets a slow-query pileup in
// cobalt exhaust the server's max_connections for everyone else. Steady-state
// usage is ~2 connections per pod, so 10 is generous.
const defaultMaxOpenConns = 10

// MaxOpenConns returns the connection pool ceiling, overridable via
// DB_MAX_OPEN_CONNS. Non-numeric or non-positive values fall back to the default.
func MaxOpenConns() int {
	if val, ok := os.LookupEnv("DB_MAX_OPEN_CONNS"); ok {
		if n, err := strconv.Atoi(val); err == nil && n > 0 {
			return n
		}
	}
	return defaultMaxOpenConns
}

func ConnectionString() string {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASS")
	database := os.Getenv("DB_NAME")

	cs := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, password, host, port, database)
	return cs
}
