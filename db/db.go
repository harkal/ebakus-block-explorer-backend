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

	var originalHash, parentHash, stateRoot, transactionsRoot, receiptsRoot []byte

	rows.Next()
	rows.Scan(&block.Number,
		&block.TimeStamp,
		&originalHash,
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

	cmpHash := strings.Join([]string{"0x", common.Bytes2Hex(originalHash)}, "")
	if strings.Compare(hash, cmpHash) != 0 {
		return nil, errors.New("wrong block found")
	}

	block.Hash = common.StringToHash(hash)
	block.ParentHash.SetBytes(parentHash)
	block.StateRoot.SetBytes(stateRoot)
	block.TransactionsRoot.SetBytes(transactionsRoot)
	block.ReceiptsRoot.SetBytes(receiptsRoot)

	return &block, nil
}

// InsertTransactions adds a number of Transactions in the database
func (cli *DBClient) InsertTransactions(txs []*models.Transaction) error {
	if len(txs) == 0 {
		return nil
	}
	log.Println("Inserting", len(txs), "transactions")
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
