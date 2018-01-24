package db

import (
	"github.com/lib/pq"
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

func (cli *DBClient) InsertBlocks(blocks []*models.Block) error {
	
	if blocks==nil || len(blocks) == 0 {
		return nil
	}

	txn, err := cli.db.Begin()
	if err != nil {
		log.Println(err.Error())
		
		return err
	}

	stmt, err := txn.Prepare(pq.CopyIn("blocks",		
		"number", 
		"timestamp", 
		"hash", 
		"parent_hash", 
		"state_root", 
		"transactions_root", 
		"receipts_root", 
		"size", 
		"gas_used",
		"gas_limit"))
	
	if err != nil {
		log.Println(err.Error())

		return err;
	}

	for _, bl := range blocks {
		_, err := stmt.Exec(
			bl.Number, 
			bl.TimeStamp, 
			bl.Hash.String(), 
			bl.ParentHash.String(), 
			bl.StateRoot.String(), 
			bl.TransactionsRoot.String(),
			bl.ReceiptsRoot.String(),
			bl.Size,
			bl.GasUsed,
			bl.GasLimit,
		)
		
		if err != nil {
			log.Println(err.Error())
		}
	}

	_, err = stmt.Exec()
	if err != nil {
		log.Println(err.Error())
	}

	err = stmt.Close()
	if err != nil {
		log.Println(err.Error())
	}

	err = txn.Commit()
	if err != nil {
		log.Println(err.Error())
	}

	return nil
}