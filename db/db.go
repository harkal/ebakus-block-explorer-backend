package db

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"text/template"

	"bitbucket.org/pantelisss/ebakus_server/models"

	"github.com/ethereum/go-ethereum/common"
	"github.com/lib/pq"
	cli "gopkg.in/urfave/cli.v1"
)

type DBClient struct {
	db *sql.DB
}

var client *DBClient

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

// InitFromCli is the same as Init but receives it's parameters
// from a Context struct of the cli package (aka from program arguments)
func InitFromCli(c *cli.Context) error {
	dbname := c.String("dbname")
	dbhost := c.String("dbhost")
	dbport := c.Int("dbport")
	dbuser := c.String("dbuser")
	dbpass := c.String("dbpass")

	return Init(dbname, dbhost, dbport, dbuser, dbpass)
}

// Init creates a connection to the database and runs any
// checks necessary to ensure the module is ready to execute
// queries.
func Init(name, host string, port int, user string, pass string) error {
	conn, err := makeConnString(name, host, port, user, pass)

	if err != nil {
		log.Println(err.Error())
		return err
	}

	fmt.Println(conn)
	tdb, err := sql.Open("postgres", conn)

	if err != nil {
		log.Println(err.Error())
		return err
	}

	err = tdb.Ping()
	if err != nil {
		log.Println(err.Error())
		return err
	}

	// Check if all required tables exist
	rows, err := tdb.Query("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'blocks');")

	if err != nil {
		log.Println(err.Error())
		return err
	}
	defer rows.Close()

	var tableExists bool
	rows.Next()
	if err := rows.Scan(&tableExists); err != nil {
		log.Println(err.Error())
		return err
	}

	if !tableExists {
		log.Println("Missing table: blocks. Make sure all required tables are created.")
		return errors.New("Missing table: blocks")
	}

	client = &DBClient{tdb}

	return nil
}

// GetClient returns the current DBClient instance.
// Dev Commentary: I'm sorry for this but I needed a way to have
// the DBClient available throughout the project. If you know
// a better way to do this I'd like to know it too.
func GetClient() *DBClient {
	return client
}

// GetLatestBlockNumber returns the most recent block id (aka number)
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

// GetBlockByID finds and returns the block with the provided ID
func (cli *DBClient) GetBlockByID(number uint64) (*models.Block, error) {
	rows, err := cli.db.Query("SELECT * FROM blocks WHERE number = $1", number)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var block models.Block

	var hash, parentHash, transactionsRoot, receiptsRoot, delegatesRaw []byte

	rows.Next()
	rows.Scan(&block.Number,
		&block.TimeStamp,
		&hash,
		&parentHash,
		&transactionsRoot,
		&receiptsRoot,
		&block.Size,
		&block.TransactionCount,
		&block.GasUsed,
		&block.GasLimit,
		&delegatesRaw)
	if err = rows.Err(); err != nil {
		return nil, err
	}

	block.Hash.SetBytes(hash)
	block.ParentHash.SetBytes(parentHash)
	block.TransactionsRoot.SetBytes(transactionsRoot)
	block.ReceiptsRoot.SetBytes(receiptsRoot)

	delegates := make([]common.Address, 0)
	l := len(delegatesRaw)
	delegateCount := l / 20
	for i := 0; i < delegateCount; i++ {
		var d common.Address
		copy(d[:], delegatesRaw[20*i:20*i+20])
		delegates = append(delegates, d)
	}

	block.Delegates = delegates

	return &block, nil
}

// GetBlockByHash finds and returns the block with the provided Hash
func (cli *DBClient) GetBlockByHash(hash string) (*models.Block, error) {
	// Query for bytea value with the hex method, pass from char [1,end) since
	// the required structure is E'\\xDEADBEEF'
	// For more, check https://www.postgresql.org/docs/9.0/static/datatype-binary.html
	query := strings.Join([]string{"SELECT * FROM blocks WHERE hash = E'\\\\", hash[1:], "'"}, "")
	rows, err := cli.db.Query(query)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var block models.Block

	var originalHash, parentHash, transactionsRoot, receiptsRoot, delegatesRaw []byte

	rows.Next()
	rows.Scan(&block.Number,
		&block.TimeStamp,
		&originalHash,
		&parentHash,
		&transactionsRoot,
		&receiptsRoot,
		&block.Size,
		&block.TransactionCount,
		&block.GasUsed,
		&block.GasLimit,
		&delegatesRaw)
	if err = rows.Err(); err != nil {
		return nil, err
	}

	cmpHash := strings.Join([]string{"0x", common.Bytes2Hex(originalHash)}, "")
	if strings.Compare(hash, cmpHash) != 0 {
		return nil, errors.New("wrong block found")
	}

	block.Hash = common.BytesToHash([]byte(hash))
	block.ParentHash.SetBytes(parentHash)
	block.TransactionsRoot.SetBytes(transactionsRoot)
	block.ReceiptsRoot.SetBytes(receiptsRoot)

	delegates := make([]common.Address, 0)
	l := len(delegatesRaw)
	delegateCount := l / 20
	for i := 0; i < delegateCount; i++ {
		var d common.Address
		copy(d[:], delegatesRaw[20*i:20*i+20])
		delegates = append(delegates, d)
	}

	block.Delegates = delegates

	return &block, nil
}

// GetBlockByID finds and returns the block with the provided ID
func (cli *DBClient) GetBlockRange(fromNumber, rng uint32) ([]models.Block, error) {
	rows, err := cli.db.Query("SELECT * FROM blocks WHERE number <= $1 order by number desc limit $2", fromNumber, rng)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.Block

	for rows.Next() {
		var block models.Block
		var hash, parentHash, transactionsRoot, receiptsRoot, delegatesRaw []byte
		rows.Scan(&block.Number,
			&block.TimeStamp,
			&hash,
			&parentHash,
			&transactionsRoot,
			&receiptsRoot,
			&block.Size,
			&block.TransactionCount,
			&block.GasUsed,
			&block.GasLimit,
			&delegatesRaw)
		if err = rows.Err(); err != nil {
			return nil, err
		}

		block.Hash.SetBytes(hash)
		block.ParentHash.SetBytes(parentHash)
		block.TransactionsRoot.SetBytes(transactionsRoot)
		block.ReceiptsRoot.SetBytes(receiptsRoot)

		delegates := make([]common.Address, 0)
		l := len(delegatesRaw)
		delegateCount := l / 20
		for i := 0; i < delegateCount; i++ {
			var d common.Address
			copy(d[:], delegatesRaw[20*i:20*i+20])
			delegates = append(delegates, d)
		}

		block.Delegates = delegates

		result = append(result, block)
	}

	return result, nil
}

// GetTransactionByHash finds and returns the transaction with the provided Hash
func (cli *DBClient) GetTransactionByHash(hash string) (*models.TransactionFull, error) {
	// Query for bytea value with the hex method, pass from char [1,end) since
	// the required structure is E'\\xDEADBEEF'
	// For more, check https://www.postgresql.org/docs/9.0/static/datatype-binary.html
	query := strings.Join([]string{"SELECT * FROM transactions WHERE hash = E'\\\\", hash[1:], "'"}, "")
	rows, err := cli.db.Query(query)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tx models.Transaction
	var txr models.TransactionReceipt

	var originalHash, blockHash, addrfrom, addrto, input []byte

	rows.Next()
	rows.Scan(&originalHash,
		&tx.Nonce,
		&blockHash,
		&tx.BlockNumber,
		&tx.TransactionIndex,
		&addrfrom,
		&addrto,
		&tx.Value,
		&tx.GasLimit,
		&txr.GasUsed,
		&tx.GasPrice,
		&input,
		&txr.Status,
		&tx.WorkNonce)
	if err = rows.Err(); err != nil {
		return nil, err
	}

	cmpHash := strings.Join([]string{"0x", common.Bytes2Hex(originalHash)}, "")
	if strings.Compare(hash, cmpHash) != 0 {
		return nil, errors.New("wrong transaction found")
	}

	tx.Hash = common.BytesToHash(originalHash)
	tx.BlockHash.SetBytes(blockHash)
	tx.From.SetBytes(addrfrom)
	tx.To.SetBytes(addrto)

	tx.Input = input

	return &models.TransactionFull{Tx: &tx, Txr: &txr}, nil
}

// GetTransactionByAddress finds and returns the transaction with the provided address
// as source (FROM) or destination (TO), or the transactions of a block
func (cli *DBClient) GetTransactionsByAddress(address string, addrtype models.AddressType) ([]models.Transaction, error) {
	// Query for bytea value with the hex method, pass from char [1,end) since
	// the required structure is E'\\xDEADBEEF'
	// For more, check https://www.postgresql.org/docs/9.0/static/datatype-binary.html
	var query string
	switch addrtype {
	case models.ADDRESS_TO:
		query = strings.Join([]string{"SELECT * FROM transactions WHERE addr_to = E'\\\\", address[1:], "'"}, "")
	case models.ADDRESS_FROM:
		query = strings.Join([]string{"SELECT * FROM transactions WHERE addr_from = E'\\\\", address[1:], "'"}, "")
	case models.ADDRESS_BLOCKHASH:
		query = strings.Join([]string{"SELECT * FROM transactions WHERE block_hash = E'\\\\", address[1:], "'"}, "")
	}

	rows, err := cli.db.Query(query)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.Transaction

	for rows.Next() {
		var tx models.Transaction
		var originalHash, blockHash, addrfrom, addrto, input []byte

		rows.Scan(&originalHash,
			&tx.Nonce,
			&blockHash,
			&tx.BlockNumber,
			&tx.TransactionIndex,
			&addrfrom,
			&addrto,
			&tx.Value,
			&tx.GasLimit,
			&input)
		if err = rows.Err(); err != nil {
			return nil, err
		}

		/*
			var cmpAddr string
			if addrtype == models.ADDRESS_TO {
				cmpAddr = strings.Join([]string{"0x", common.Bytes2Hex(addrto)}, "")
			} else {
				cmpAddr = strings.Join([]string{"0x", common.Bytes2Hex(addrfrom)}, "")
			}
			if strings.Compare(address, cmpAddr) != 0 {
				return nil, errors.New("wrong transaction found")
			}
		*/

		tx.Hash = common.BytesToHash(originalHash)
		tx.BlockHash.SetBytes(blockHash)
		tx.From.SetBytes(addrfrom)
		tx.To.SetBytes(addrto)

		result = append(result, tx)
	}

	return result, nil
}

// InsertTransactions adds a number of Transactions in the database
func (cli *DBClient) InsertTransactions(transactions []models.TransactionFull) error {
	if len(transactions) == 0 {
		return nil
	}

	txn, err := cli.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := txn.Prepare(pq.CopyIn("transactions",
		"hash",
		"status",
		"nonce",
		"block_hash",
		"block_number",
		"tx_index",
		"addr_from",
		"addr_to",
		"value",
		"gasused",
		"gaslimit",
		"gasprice",
		"worknonce",
		"input"))

	if err != nil {
		return err
	}

	for _, txf := range transactions {
		tx := txf.Tx
		txr := txf.Txr
		log.Println("Adding", tx.BlockNumber, tx.TransactionIndex, tx.Input)

		v := uint64(tx.Value >> 1) // stupid go postgres driver
		_, err := stmt.Exec(
			tx.Hash.Bytes(),
			txr.Status,
			tx.Nonce,
			tx.BlockHash.Bytes(),
			tx.BlockNumber,
			tx.TransactionIndex,
			tx.From.Bytes(),
			tx.To.Bytes(),
			v,
			txr.GasUsed,
			tx.GasLimit,
			tx.GasPrice,
			tx.WorkNonce,
			tx.Input,
		)

		if err != nil {
			log.Println("Error on Transaction", tx.BlockNumber, err.Error())
		}
	}

	_, err = stmt.Exec()
	if err != nil {
		log.Println("PQTX Exec", err.Error())
	}

	err = stmt.Close()
	if err != nil {
		log.Println("PQTX Close", err.Error())
	}

	err = txn.Commit()
	if err != nil {
		log.Println("PQTX Commit", err.Error())
	}

	return nil
}

// InsertBlocks adds a number of Blocks in the database
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
		"transactions_root",
		"receipts_root",
		"size",
		"transaction_count",
		"gas_used",
		"gas_limit",
		"delegates"))

	if err != nil {
		return err
	}

	for _, bl := range blocks {
		dbytes := make([]byte, 0)
		for _, d := range bl.Delegates {
			dbytes = append(dbytes, d[:]...)
		}
		_, err := stmt.Exec(
			bl.Number,
			bl.TimeStamp,
			bl.Hash.Bytes(),
			bl.ParentHash.Bytes(),
			bl.TransactionsRoot.Bytes(),
			bl.ReceiptsRoot.Bytes(),
			bl.Size,
			len(bl.Transactions),
			bl.GasUsed,
			bl.GasLimit,
			dbytes,
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
