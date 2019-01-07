package db

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strings"
	"text/template"

	"bitbucket.org/pantelisss/ebakus_server/models"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/lib/pq"
	cli "gopkg.in/urfave/cli.v1"
)

type DBClient struct {
	db *sql.DB
}

var client *DBClient

var (
	valueDecimalPoints = int64(4)
	precisionFactor    = new(big.Int).Exp(big.NewInt(10), big.NewInt(18-valueDecimalPoints), nil)
)

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

	var hash, parentHash, transactionsRoot, receiptsRoot, delegatesRaw, producer []byte

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
		&delegatesRaw,
		&producer,
		&block.Signature)
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
	block.Producer.SetBytes(producer)

	return &block, nil
}

func (cli *DBClient) ScanBlock() {

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

	var originalHash, parentHash, transactionsRoot, receiptsRoot, delegatesRaw, producer []byte

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
		&delegatesRaw,
		&producer,
		&block.Signature)
	if err = rows.Err(); err != nil {
		return nil, err
	}

	cmpHash := strings.Join([]string{"0x", common.Bytes2Hex(originalHash)}, "")
	if strings.Compare(hash, cmpHash) != 0 {
		return nil, errors.New("wrong block found")
	}

	block.Hash.SetBytes(originalHash)
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
	block.Producer.SetBytes(producer)

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
		var hash, parentHash, transactionsRoot, receiptsRoot, delegatesRaw, producer []byte
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
			&delegatesRaw,
			&producer,
			&block.Signature)
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
		block.Producer.SetBytes(producer)

		result = append(result, block)
	}

	return result, nil
}

// GetTransactionRange finds and returns the transactions in a specific range
func (cli *DBClient) GetTransactionRange(hash string, rng uint32) ([]models.TransactionFull, error) {

	var query string
	// get latest txs
	if hash == "-1" {
		query = "SELECT * FROM transactions ORDER BY timestamp DESC LIMIT $1"

		// get latest txs after hash
	} else {
		query = strings.Join([]string{"SELECT * FROM transactions WHERE timestamp >= (SELECT timestamp FROM transactions WHERE hash = E'\\\\", hash[1:], "') ORDER BY timestamp DESC LIMIT $1"}, "")
	}

	rows, err := cli.db.Query(query, rng)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.TransactionFull

	for rows.Next() {
		var tx models.Transaction
		var txr models.TransactionReceipt

		var originalHash, blockHash, addrfrom, addrto, input []byte
		var value uint64

		rows.Scan(&originalHash,
			&tx.Nonce,
			&blockHash,
			&tx.BlockNumber,
			&tx.TransactionIndex,
			&addrfrom,
			&addrto,
			&value,
			&tx.GasLimit,
			&txr.GasUsed,
			&txr.CumulativeGasUsed,
			&tx.GasPrice,
			&input,
			&txr.Status,
			&tx.WorkNonce,
			&tx.Timestamp)
		if err = rows.Err(); err != nil {
			return nil, err
		}

		tx.Hash = common.BytesToHash(originalHash)
		tx.BlockHash.SetBytes(blockHash)
		tx.From.SetBytes(addrfrom)
		tx.To.SetBytes(addrto)
		tx.Value = (hexutil.Big)(*new(big.Int).Mul(new(big.Int).SetUint64(value), precisionFactor)) // value * ether (1e18) / 10000

		tx.Input = input

		txf := models.TransactionFull{Tx: &tx, Txr: &txr}
		result = append(result, txf)
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
	var value uint64

	rows.Next()
	rows.Scan(&originalHash,
		&tx.Nonce,
		&blockHash,
		&tx.BlockNumber,
		&tx.TransactionIndex,
		&addrfrom,
		&addrto,
		&value,
		&tx.GasLimit,
		&txr.GasUsed,
		&txr.CumulativeGasUsed,
		&tx.GasPrice,
		&input,
		&txr.Status,
		&tx.WorkNonce,
		&tx.Timestamp)
	if err = rows.Err(); err != nil {
		return nil, err
	}

	tx.Hash = common.BytesToHash(originalHash)
	tx.BlockHash.SetBytes(blockHash)
	tx.From.SetBytes(addrfrom)
	tx.To.SetBytes(addrto)
	tx.Value = (hexutil.Big)(*new(big.Int).Mul(new(big.Int).SetUint64(value), precisionFactor)) // value * ether (1e18) / 10000

	tx.Input = input

	return &models.TransactionFull{Tx: &tx, Txr: &txr}, nil
}

func (cli *DBClient) GetAddressTotals(address string) (sumIn, sumOut float64, countIn, countOut uint64, err error) {

	query := strings.Join([]string{"SELECT count(value), sum(value) FROM transactions WHERE addr_to = E'\\\\", address[1:], "'"}, "")

	rows, err := cli.db.Query(query)

	if err != nil {
		return 0, 0, 0, 0, err
	}
	defer rows.Close()

	var sumInEbakus float64

	rows.Next()
	rows.Scan(&countIn,
		&sumInEbakus)
	if err = rows.Err(); err != nil {
		return 0, 0, 0, 0, err
	}

	query = strings.Join([]string{"SELECT count(value), sum(value) FROM transactions WHERE addr_from = E'\\\\", address[1:], "'"}, "")

	rows, err = cli.db.Query(query)

	if err != nil {
		return 0, 0, 0, 0, err
	}
	defer rows.Close()

	var sumOutEbakus float64

	rows.Next()
	rows.Scan(&countOut, &sumOutEbakus)
	if err = rows.Err(); err != nil {
		return 0, 0, 0, 0, err
	}

	sumIn = sumInEbakus * 2.0
	sumOut = sumOutEbakus * 2.0

	return
}

// GetTransactionByAddress finds and returns the transaction with the provided address
// as source (FROM) or destination (TO), or the transactions of a block
func (cli *DBClient) GetTransactionsByAddress(address string, addrtype models.AddressType, offset, limit uint64, order string) ([]models.TransactionFull, error) {
	// Query for bytea value with the hex method, pass from char [1,end) since
	// the required structure is E'\\xDEADBEEF'
	// For more, check https://www.postgresql.org/docs/9.0/static/datatype-binary.html
	var query string
	switch addrtype {
	case models.ADDRESS_TO:
		query = strings.Join([]string{"SELECT * FROM transactions WHERE addr_to = E'\\\\", address[1:], "'"}, "")
	case models.ADDRESS_FROM:
		query = strings.Join([]string{"SELECT * FROM transactions WHERE addr_from = E'\\\\", address[1:], "'"}, "")
	case models.ADDRESS_ALL:
		query = strings.Join([]string{"SELECT * FROM transactions WHERE addr_to = E'\\\\", address[1:], "'", " or addr_from = E'\\\\", address[1:], "'"}, "")
	case models.ADDRESS_BLOCKHASH:
		query = strings.Join([]string{"SELECT * FROM transactions WHERE block_hash = E'\\\\", address[1:], "'"}, "")
	}

	if order != "asc" {
		query = strings.Join([]string{query, " ORDER BY timestamp ", order}, "")
	}

	query = strings.Join([]string{query, " OFFSET $1 LIMIT $2"}, "")

	rows, err := cli.db.Query(query, offset, limit)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.TransactionFull

	result = make([]models.TransactionFull, 0)

	for rows.Next() {
		var tx models.Transaction
		var txr models.TransactionReceipt

		var originalHash, blockHash, addrfrom, addrto, input []byte
		var value uint64

		rows.Scan(&originalHash,
			&tx.Nonce,
			&blockHash,
			&tx.BlockNumber,
			&tx.TransactionIndex,
			&addrfrom,
			&addrto,
			&value,
			&tx.GasLimit,
			&txr.GasUsed,
			&txr.CumulativeGasUsed,
			&tx.GasPrice,
			&input,
			&txr.Status,
			&tx.WorkNonce,
			&tx.Timestamp)
		if err = rows.Err(); err != nil {
			log.Println(err)
			return nil, err
		}

		tx.Hash = common.BytesToHash(originalHash)
		tx.BlockHash.SetBytes(blockHash)
		tx.From.SetBytes(addrfrom)
		tx.To.SetBytes(addrto)
		tx.Value = (hexutil.Big)(*new(big.Int).Mul(new(big.Int).SetUint64(value), precisionFactor)) // value * ether (1e18) / 10000

		tx.Input = input

		result = append(result, models.TransactionFull{Tx: &tx, Txr: &txr})
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
		"timestamp",
		"status",
		"nonce",
		"block_hash",
		"block_number",
		"tx_index",
		"addr_from",
		"addr_to",
		"value",
		"gasused",
		"cumulativegasused",
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

		// value * 10000 / ether (1e18)
		v := new(big.Int).Div(tx.Value.ToInt(), precisionFactor).Uint64() // stupid go postgres driver
		_, err := stmt.Exec(
			tx.Hash.Bytes(),
			tx.Timestamp,
			txr.Status,
			tx.Nonce,
			tx.BlockHash.Bytes(),
			tx.BlockNumber,
			tx.TransactionIndex,
			tx.From.Bytes(),
			tx.To.Bytes(),
			v,
			txr.GasUsed,
			txr.CumulativeGasUsed,
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
		"delegates",
		"producer",
		"signature"))

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
			bl.Producer,
			bl.Signature,
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
