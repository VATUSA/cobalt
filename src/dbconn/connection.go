package dbconn

import (
	"database/sql"
	"log"
	"vatusa-cobalt/config"
	"vatusa-cobalt/db"

	_ "github.com/go-sql-driver/mysql"
)

var database *sql.DB

func DB() *sql.DB {
	if database == nil {
		var err error
		database, err = sql.Open("mysql", config.ConnectionString())
		if err != nil {
			log.Fatal(err)
		}
		database.SetMaxOpenConns(config.MaxOpenConns())
	}
	return database
}

func Queries() *db.Queries {
	return db.New(DB())
}
