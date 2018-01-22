package db

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq" // Register some standard stuff
)

const (
	host   = "localhost"
	port   = 5432
	user   = "postgres"
	dbname = "postgres"
)

var db *sql.DB	


func init() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"dbname=%s sslmode=disable",
		host, port, user, dbname)
	tdb, err := sql.Open("postgres", psqlInfo)
	db = tdb
	if err != nil {
		panic(err)
	}
	// defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected!")
}
