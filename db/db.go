package db

import (
	"database/sql"
	"fmt"
	"log"

	"bitbucket.org/pantelisss/ebakus_server/models"

	"github.com/lib/pq"
)

type DBClient struct {
	db *sql.DB
}

func NewClient(name, host string, port int, user string, pass string) (*DBClient, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"dbname=%s sslmode=disable",
		host, port, user, name)
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

func (cli *DBClient) GetBlock(number uint64) (*models.Block, error) {
	rows, err := cli.db.Query("SELECT * FROM blocks WHERE number = $1", number)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var block models.Block

	var hash, parentHash, stateRoot, transactionsRoot, receiptsRoot []byte

	rows.Next()
	rows.Scan(&block.Number,
		&block.TimeStamp,
		&hash,
		&parentHash,
		&stateRoot,
		&transactionsRoot,
		&receiptsRoot,
		&block.Size,
		&block.GasUsed,
		&block.GasLimit)
	if err = rows.Err(); err != nil {
		return nil, err
	}

	block.Hash.SetBytes(hash)
	block.ParentHash.SetBytes(parentHash)
	block.StateRoot.SetBytes(stateRoot)
	block.TransactionsRoot.SetBytes(transactionsRoot)
	block.ReceiptsRoot.SetBytes(receiptsRoot)

	return &block, nil
}

func (cli *DBClient) InsertTransactions(txs []*models.Transaction) error {
	if len(txs) == 0 {
		return nil
	}
	log.Println("Insert transaction", len(txs))
	return nil
}

func (cli *DBClient) InsertBlocks(blocks []*models.Block) error {
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
