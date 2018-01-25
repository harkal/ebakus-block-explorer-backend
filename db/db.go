package db

import (
	"database/sql"
	"ebakus_server/models"
	"fmt"
	"log"

	"github.com/lib/pq"

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

	err = tdb.Ping()
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	return &DBClient{tdb}, nil
}

func (cli *DBClient) GetLatestBlockNumber() (uint64, error) {
	rows, err := cli.db.Query("SELECT max(number) FROM blocks")
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var maxNumber uint64

	rows.Next()
	rows.Scan(&maxNumber)
	if err = rows.Err(); err != nil {
		return 0, err
	}

	return maxNumber, nil
}

func (cli *DBClient) InsertBlocks(blocks []models.Block) error {
	if len(blocks) == 0 {
		return nil
	}

	txn, err := cli.db.Begin()
	if err != nil {
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
		return err
	}

	for _, bl := range blocks {
		_, err := stmt.Exec(
			bl.Number,
			bl.TimeStamp,
			bl.Hash.Bytes(),
			bl.ParentHash.Bytes(),
			bl.StateRoot.Bytes(),
			bl.TransactionsRoot.Bytes(),
			bl.ReceiptsRoot.Bytes(),
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
