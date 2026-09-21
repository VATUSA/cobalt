package config

import (
	"fmt"
	"net/url"
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

// ConnectionString builds the go-sql-driver/mysql DSN.
//
// TLS is opt-in through DB_TLS, which maps straight onto the driver's `tls`
// parameter: "true" encrypts and fully verifies the server certificate and
// hostname against the system roots, "skip-verify" encrypts without verifying,
// "preferred" uses TLS only if the server offers it.
//
// It is opt-in rather than on by default because the two environments differ.
// Azure Database for MySQL sets require_secure_transport=ON and presents a
// certificate chaining to DigiCert Global Root G2 with the server hostname in
// its SANs, so "true" verifies cleanly against the CA bundle already in the
// image. DigitalOcean's managed MySQL is reached over a private VPC endpoint
// and presents a certificate from DigitalOcean's own CA, which is not in the
// system roots — defaulting to "true" would break every existing deployment,
// including production.
//
// Leaving DB_TLS unset omits the parameter entirely, which is byte-identical to
// the DSN this function produced before TLS was configurable.
func ConnectionString() string {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASS")
	database := os.Getenv("DB_NAME")

	cs := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, password, host, port, database)
	if tls := os.Getenv("DB_TLS"); tls != "" {
		cs += "&tls=" + url.QueryEscape(tls)
	}
	return cs
}
