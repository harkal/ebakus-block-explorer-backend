package db

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"text/template"

	"bitbucket.org/pantelisss/ebakus_server/models"

	"github.com/lib/pq"
)

type DBClient struct {
	db *sql.DB
}

func makeConnString(name, host string, port int, user string, pass string) (string, error) {
	templ, err := template.New("psql_connection_string").Parse("postgres://{{.User}}:{{.Pass}}@{{.Host}}:{{.Port}}/{{.Name}}?sslmode=disable")

	if err != nil {
		log.Println(err.Error())
		return string(""), err
	}

	data := struct {
		User string
		Pass string
		Host string
		Port int
		Name string
	}{
		user,
		pass,
		host,
		port,
		name,
	}

	buff := new(bytes.Buffer)
	err = templ.Execute(buff, data)

	return buff.String(), err
}

func NewClient(name, host string, port int, user string, pass string) (*DBClient, error) {
	conn, err := makeConnString(name, host, port, user, pass)

	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	fmt.Println(conn)
	tdb, err := sql.Open("postgres", conn)

	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	err = tdb.Ping()
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	// Check if all required tables exist
	rows, err := tdb.Query("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'blocks');")

	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	defer rows.Close()

	var tableExists bool
	rows.Next()
	if err := rows.Scan(&tableExists); err != nil {
		log.Println(err.Error())
		return nil, err
	}

	if !tableExists {
		log.Println("Missing table: blocks. Make sure all required tables are created.")
		return nil, errors.New("Missing table: blocks")
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
