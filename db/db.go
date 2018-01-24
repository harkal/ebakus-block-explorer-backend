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

func (cli *DBClient) InsertBlocks(blocks []*models.Block) error {
	stmt := `INSERT INTO blocks (
		number, 
		timestamp, 
		hash, 
		parent_hash, 
		state_root, 
		transactions_root, 
		receipts_root, 
		size, 
		gas_used,
		gas_limit)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	
	if blocks==nil || len(blocks) == 0 {
		return nil
	}

	bl := blocks[0]

	log.Println(		
		bl.Number, 
		bl.TimeStamp, 
		bl.Hash.String(), 
		bl.ParentHash.String(), 
		bl.StateRoot.String(), 
		bl.TransactionsRoot.String(),
		bl.ReceiptsRoot.String(),
		bl.Size,
		bl.GasUsed,
		bl.GasLimit)

	_, err := cli.db.Exec(stmt,
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
		 return err
	 }

	 return nil
}