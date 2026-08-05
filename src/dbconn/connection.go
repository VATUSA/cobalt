package dbconn

import (
	"database/sql"
	"log"
	"vatusa-cobalt/config"
	"vatusa-cobalt/db"

	_ "github.com/go-sql-driver/mysql"
)

var queries *db.Queries

func Queries() *db.Queries {
	if queries == nil {
		database, err := sql.Open("mysql", config.ConnectionString())
		if err != nil {
			log.Fatal(err)
		}
		database.SetMaxOpenConns(config.MaxOpenConns())
		queries = db.New(database)
	}
	return queries
}
