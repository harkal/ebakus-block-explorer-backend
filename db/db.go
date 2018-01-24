package db

import (
	"database/sql"
	"log"
	"fmt"
	"ebakus_server/models"

	_ "github.com/lib/pq" // Register some standard stuff
)

const (
	host   = "localhost"
	port   = 5432
	user   = "postgres"
	dbname = "ebakus"
)

type DBClient struct {
	db *sql.DB	
}


func NewClient() (*DBClient, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"dbname=%s sslmode=disable",
		host, port, user, dbname)
	tdb, err := sql.Open("postgres", psqlInfo)
	
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	// defer db.Close()

	err = tdb.Ping()
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	log.Println("Successfully connected!")

	return &DBClient{tdb}, nil
}