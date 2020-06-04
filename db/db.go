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

	"github.com/ebakus/ebakus-block-explorer-backend/models"
	"github.com/ebakus/ebakus-block-explorer-backend/redis"

	"github.com/ebakus/go-ebakus/common"
	"github.com/ebakus/go-ebakus/common/hexutil"
	"github.com/lib/pq"
	"github.com/urfave/cli"
)

type DBClient struct {
	db *sql.DB
}

var client *DBClient

var (
	valueDecimalPoints = int64(4)
	precisionFactor    = new(big.Int).Exp(big.NewInt(10), big.NewInt(18-valueDecimalPoints), nil)
	bigIntZero         = new(big.Int).SetUint64(0)
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
	query := strings.Join([]string{
		"SELECT b.*, ARRAY_TO_STRING((ARRAY_AGG(DISTINCT ens.name))[:1], ', ') producer_ens",
		" FROM blocks AS b",
		"   LEFT JOIN ens ON ens.address = b.producer",
		" WHERE b.number = $1",
		" GROUP BY b.number"}, "")
	rows, err := cli.db.Query(query, number)
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
		&block.Signature,
		&block.ProducerEns)
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
	query := strings.Join([]string{
		"SELECT b.*, ARRAY_TO_STRING((ARRAY_AGG(DISTINCT ens.name))[:1], ', ') producer_ens",
		" FROM blocks AS b",
		"   LEFT JOIN ens ON ens.address = b.producer",
		" WHERE b.hash = E'\\\\", hash[1:], "'",
		" GROUP BY b.number"}, "")
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
		&block.Signature,
		&block.ProducerEns)
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

// GetBlockRange returns range of blocks
func (cli *DBClient) GetBlockRange(fromNumber, rng uint32) ([]models.Block, error) {

	withQuery := "SELECT * FROM blocks WHERE number <= $1 ORDER BY number DESC LIMIT $2"

	query := strings.Join([]string{
		"WITH b AS (", withQuery, ")",
		" SELECT b.*,",
		"   ARRAY_TO_STRING((ARRAY_AGG(DISTINCT ens.name))[:1], ', ') producer_ens",
		" FROM b",
		"   LEFT JOIN ens ON ens.address = b.producer",
		" GROUP BY b.number, b.timestamp, b.hash, b.parent_hash, b.transactions_root, b.receipts_root,",
		"   b.size, b.transaction_count, b.gas_used, b.gas_limit, b.delegates, b.producer, b.signature",
		" ORDER BY b.number DESC LIMIT $2"}, "")

	rows, err := cli.db.Query(query, fromNumber, rng)
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
			&block.Signature,
			&block.ProducerEns)
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

// GetBlocksByTimestamp finds and returns the block info ordered by timestamp
func (cli *DBClient) GetBlocksByTimestamp(timestamp hexutil.Uint64, timestampCondition models.TimestampCondition, producer string) ([]models.Block, error) {
	// Query for bytea value with the hex method, pass from char [1,end) since
	// the required structure is E'\\xDEADBEEF'
	// For more, check https://www.postgresql.org/docs/9.0/static/datatype-binary.html
	withQuery := "SELECT * FROM blocks"

	switch timestampCondition {
	case models.TIMESTAMP_EQUAL:
		withQuery = strings.Join([]string{withQuery, " WHERE timestamp = $1"}, "")
	case models.TIMESTAMP_GREATER_EQUAL_THAN:
		withQuery = strings.Join([]string{withQuery, " WHERE timestamp >= $1"}, "")
	case models.TIMESTAMP_SMALLER_EQUAL_THAN:
		withQuery = strings.Join([]string{withQuery, " WHERE timestamp <= $1"}, "")
	}

	if common.IsHexAddress(producer) {
		withQuery = strings.Join([]string{withQuery, " AND producer = E'\\\\", producer[1:], "'"}, "")
	}

	withQuery = strings.Join([]string{withQuery, " ORDER BY timestamp DESC"}, "")

	query := strings.Join([]string{
		"WITH b AS (", withQuery, ")",
		" SELECT b.*,",
		"   ARRAY_TO_STRING((ARRAY_AGG(DISTINCT ens.name))[:1], ', ') producer_ens",
		" FROM b",
		"   LEFT JOIN ens ON ens.address = b.producer",
		" GROUP BY b.number, b.timestamp, b.hash, b.parent_hash, b.transactions_root, b.receipts_root,",
		"   b.size, b.transaction_count, b.gas_used, b.gas_limit, b.delegates, b.producer, b.signature",
		" ORDER BY b.timestamp DESC"}, "")

	rows, err := cli.db.Query(query, timestamp)
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
			&block.Signature,
			&block.ProducerEns)
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

// GetTransactionByHash finds and returns the transaction with the provided Hash
func (cli *DBClient) GetTransactionByHash(hash string) (*models.TransactionFull, error) {
	// Query for bytea value with the hex method, pass from char [1,end) since
	// the required structure is E'\\xDEADBEEF'
	// For more, check https://www.postgresql.org/docs/9.0/static/datatype-binary.html

	withQuery := strings.Join([]string{"SELECT * FROM transactions WHERE hash = E'\\\\", hash[1:], "'"}, "")

	query := strings.Join([]string{
		"WITH t AS (", withQuery, ")",
		" SELECT t.*,",
		"   ARRAY_TO_STRING((ARRAY_AGG(DISTINCT ensf.name))[:1], ', ') from_ens,",
		"   ARRAY_TO_STRING((ARRAY_AGG(DISTINCT enst.name))[:1], ', ') to_ens,",
		"   ARRAY_TO_STRING((ARRAY_AGG(DISTINCT ensc.name))[:1], ', ') contract_ens",
		" FROM t",
		"   LEFT JOIN ens AS ensf ON ensf.address = t.addr_from",
		"   LEFT JOIN ens AS enst ON enst.address = t.addr_to",
		"   LEFT JOIN ens AS ensc ON ensc.address = t.contract_address",
		" GROUP BY t.hash, t.nonce, t.block_hash, t.block_number, t.tx_index, t.addr_from, t.addr_to, t.value,",
		"   t.gas_limit, t.gas_used, t.cumulative_gas_used, t.gas_price, t.contract_address, t.input, t.status,",
		"   t.work_nonce, t.timestamp"}, "")
	rows, err := cli.db.Query(query)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tx models.Transaction
	var txr models.TransactionReceipt

	var originalHash, blockHash, addrfrom, addrto, addrContract, input []byte
	var value uint64

	if foundData := rows.Next(); !foundData {
		return &models.TransactionFull{Tx: nil, Txr: nil}, nil
	}

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
		&addrContract,
		&input,
		&txr.Status,
		&tx.WorkNonce,
		&tx.Timestamp,
		&tx.FromEns,
		&tx.ToEns,
		&txr.ContractAddressEns)
	if err = rows.Err(); err != nil {
		return nil, err
	}

	tx.Hash = common.BytesToHash(originalHash)
	tx.BlockHash.SetBytes(blockHash)
	tx.From.SetBytes(addrfrom)
	addressTo := common.BytesToAddress(addrto)
	tx.To = &addressTo
	tx.Value = (hexutil.Big)(*new(big.Int).Mul(new(big.Int).SetUint64(value), precisionFactor)) // value * ether (1e18) / 10000

	contractAddress := common.BytesToAddress(addrContract)
	txr.ContractAddress = &contractAddress
	tx.Input = input

	return &models.TransactionFull{Tx: &tx, Txr: &txr}, nil
}

// DeleteTransactionsAndBlockByID deletes the block and its transactions by block number
func (cli *DBClient) DeleteBlockWithTransactionsByID(number uint64, producer common.Address) (err error) {
	txn, err := cli.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			// a panic occurred, rollback and repanic
			txn.Rollback()
			panic(p)
		} else if err != nil {
			txn.Rollback()
		} else {
			err = txn.Commit()
		}
	}()

	_, err = txn.Exec("DELETE FROM blocks WHERE number = $1", number)
	if err != nil {
		return err
	}

	_, err = txn.Exec("DELETE FROM transactions WHERE block_number = $1", number)
	if err != nil {
		return err
	}

	sql := `
		INSERT INTO producers(address) VALUES (E'\\x%s')
		ON CONFLICT (address) DO UPDATE
			SET produced_blocks_count = producers.produced_blocks_count - 1,
				block_rewards = producers.block_rewards - 3171
			WHERE producers.produced_blocks_count > 0 AND producers.block_rewards >= 3171
	`
	producerAddr := common.Bytes2Hex(producer[:])[:]
	if _, err := txn.Query(fmt.Sprintf(sql, producerAddr)); err != nil {
		return err
	}

	return err
}

func (cli *DBClient) GetAddressTotals(address string) (blockRewards *big.Int, txCount uint64, err error) {

	query := strings.Join([]string{"SELECT count(*) FROM transactions WHERE addr_from = E'\\\\", address[1:], "' OR addr_to = E'\\\\", address[1:], "'"}, "")
	rows, err := cli.db.Query(query)

	if err != nil {
		return bigIntZero, 0, err
	}
	defer rows.Close()

	rows.Next()
	rows.Scan(&txCount)
	if err = rows.Err(); err != nil {
		return bigIntZero, 0, err
	}

	producer, err := cli.GetProducer(address)
	if err != nil {
		return bigIntZero, txCount, err
	}

	blockRewards = producer.BlockRewards

	return
}

// GetIsContractAddress checks if an address is a contract
func (cli *DBClient) GetIsContractAddress(address string) (bool, error) {
	// return true for system contracts
	if bytes.Compare(common.HexToAddress(address).Bytes(), []byte{1, 2}) <= 0 {
		return true, nil
	}

	query := strings.Join([]string{"SELECT count(*) FROM transactions WHERE contract_address = E'\\\\", address[1:], "'"}, "")
	varInt := uint64(0)
	cli.db.QueryRow(query).Scan(&varInt)
	return varInt > 0, nil
}

// GetTransactionByAddress finds and returns the transaction with the provided address
// as source (FROM) or destination (TO), or the transactions of a block
func (cli *DBClient) GetTransactionsByAddress(address string, addrtype models.AddressType, offset, limit uint64, order string) ([]models.TransactionFull, error) {
	// Query for bytea value with the hex method, pass from char [1,end) since
	// the required structure is E'\\xDEADBEEF'
	// For more, check https://www.postgresql.org/docs/9.0/static/datatype-binary.html
	withQuery := "SELECT * FROM transactions"

	switch addrtype {
	case models.ADDRESS_TO:
		withQuery = strings.Join([]string{withQuery, " WHERE addr_to = E'\\\\", address[1:], "'"}, "")
	case models.ADDRESS_FROM:
		withQuery = strings.Join([]string{withQuery, " WHERE addr_from = E'\\\\", address[1:], "'"}, "")
	case models.ADDRESS_ALL:
		withQuery = strings.Join([]string{withQuery, " WHERE addr_to = E'\\\\", address[1:], "'", " or addr_from = E'\\\\", address[1:], "'"}, "")
	case models.ADDRESS_BLOCKHASH:
		withQuery = strings.Join([]string{"SELECT transactions.* FROM transactions, blocks WHERE blocks.number = transactions.block_number AND blocks.hash = E'\\\\", address[1:], "'"}, "")
	}

	if order != "asc" {
		switch addrtype {
		case models.LATEST:
			withQuery = strings.Join([]string{withQuery, " ORDER BY block_number ", order}, "")
		default:
			withQuery = strings.Join([]string{withQuery, " ORDER BY timestamp ", order}, "")
		}
	}

	withQuery = strings.Join([]string{withQuery, " OFFSET $1 LIMIT $2"}, "")

	query := strings.Join([]string{
		"WITH t AS (", withQuery, ")",
		" SELECT t.*,",
		"   ARRAY_TO_STRING((ARRAY_AGG(DISTINCT ensf.name))[:1], ', ') from_ens,",
		"   ARRAY_TO_STRING((ARRAY_AGG(DISTINCT enst.name))[:1], ', ') to_ens,",
		"   ARRAY_TO_STRING((ARRAY_AGG(DISTINCT ensc.name))[:1], ', ') contract_ens",
		" FROM t",
		"   LEFT JOIN ens AS ensf ON ensf.address = t.addr_from",
		"   LEFT JOIN ens AS enst ON enst.address = t.addr_to",
		"   LEFT JOIN ens AS ensc ON ensc.address = t.contract_address",
		" GROUP BY t.hash, t.nonce, t.block_hash, t.block_number, t.tx_index, t.addr_from, t.addr_to,",
		"   t.value, t.gas_limit, t.gas_used, t.cumulative_gas_used, t.gas_price, t.contract_address,",
		"   t.input, t.status, t.work_nonce, t.timestamp"}, "")

	if order != "asc" {
		switch addrtype {
		case models.LATEST:
			query = strings.Join([]string{query, " ORDER BY t.block_number ", order}, "")
		default:
			query = strings.Join([]string{query, " ORDER BY t.timestamp ", order}, "")
		}
	}

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

		var originalHash, blockHash, addrfrom, addrto, addrContract, input []byte
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
			&addrContract,
			&input,
			&txr.Status,
			&tx.WorkNonce,
			&tx.Timestamp,
			&tx.FromEns,
			&tx.ToEns,
			&txr.ContractAddressEns)
		if err = rows.Err(); err != nil {
			log.Println(err)
			return nil, err
		}

		tx.Hash = common.BytesToHash(originalHash)
		tx.BlockHash.SetBytes(blockHash)
		tx.From.SetBytes(addrfrom)
		addressTo := common.BytesToAddress(addrto)
		tx.To = &addressTo
		tx.Value = (hexutil.Big)(*new(big.Int).Mul(new(big.Int).SetUint64(value), precisionFactor)) // value * ether (1e18) / 10000

		contractAddress := common.BytesToAddress(addrContract)
		txr.ContractAddress = &contractAddress
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
		"gas_used",
		"cumulative_gas_used",
		"gas_limit",
		"gas_price",
		"work_nonce",
		"contract_address",
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
		var to, contractAddress []byte
		if tx.To != nil {
			to = tx.To.Bytes()
		}
		if txr.ContractAddress != nil {
			contractAddress = txr.ContractAddress.Bytes()
		}

		_, err := stmt.Exec(
			tx.Hash.Bytes(),
			tx.Timestamp,
			txr.Status,
			tx.Nonce,
			tx.BlockHash.Bytes(),
			tx.BlockNumber,
			tx.TransactionIndex,
			tx.From.Bytes(),
			to,
			v,
			txr.GasUsed,
			txr.CumulativeGasUsed,
			tx.GasLimit,
			tx.GasPrice,
			tx.WorkNonce,
			contractAddress,
			tx.Input,
		)

		if err != nil {
			log.Println("Error on Transaction", tx.BlockNumber, err.Error())
		}

		if tx.To != nil {
			if err := redis.Delete("address:" + tx.To.Hex()); err != nil {
				log.Println("Failed to clear redis cache for ", "address:"+tx.To.Hex(), err.Error())
			}
		}

		if err := redis.Delete("address:" + tx.From.Hex()); err != nil {
			log.Println("Failed to clear redis cache for ", "address:"+tx.From.Hex(), err.Error())
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

		if err := redis.Delete("address:" + bl.Producer.Hex()); err != nil {
			log.Println("Failed to clear redis cache for ", "address:"+bl.Producer.Hex(), err.Error())
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

// InsertBalance inserts/updates the balance of an address
func (cli *DBClient) InsertBalance(address common.Address, liquidBalance uint64, stakedBalance uint64, blockNumber uint64) error {
	sql := `
		INSERT INTO balances(address, liquid_amount, staked_amount, block_number) VALUES (E'\\x%s', %d, %d, %d)
		ON CONFLICT (address) DO UPDATE
			SET liquid_amount = excluded.liquid_amount,
				staked_amount = excluded.staked_amount,
				block_number = excluded.block_number
	`
	adr := common.Bytes2Hex(address[:])[:]
	//	log.Println(fmt.Sprintf(sql, adr, liquidBalance, stakedBalance, blockNumber))
	rows, err := cli.db.Query(fmt.Sprintf(sql, adr, liquidBalance, stakedBalance, blockNumber))
	rows.Close()

	return err
}

// GetBalanceStats gets the table stats
func (cli *DBClient) GetBalanceStats() (uint64, uint64, uint64, uint64, uint64, error) {
	query := `select count(*), max(liquid_amount), min(liquid_amount), max(staked_amount), min(staked_amount) from balances`
	var count, maxLiquid, minLiquid, maxStaked, minStaked uint64
	err := cli.db.QueryRow(query).Scan(&count, &maxLiquid, &minLiquid, &maxStaked, &minStaked)
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}

	return count, maxLiquid, minLiquid, maxStaked, minStaked, nil
}

// GetTopBalances gets the rich list
func (cli *DBClient) GetTopBalances(limit uint64, offset uint64) ([]models.Balance, error) {
	withQuery := "SELECT address, liquid_amount, staked_amount, block_number FROM balances ORDER BY liquid_amount + staked_amount DESC LIMIT $1 OFFSET $2"

	query := strings.Join([]string{
		"WITH b AS (", withQuery, ")",
		" SELECT b.*,",
		"   ARRAY_TO_STRING((ARRAY_AGG(DISTINCT ens.name))[:1], ', ') address_ens",
		" FROM b",
		"   LEFT JOIN ens ON ens.address = b.address",
		" GROUP BY b.address, b.liquid_amount, b.staked_amount, b.block_number",
		" ORDER BY b.liquid_amount + b.staked_amount DESC"}, "")
	rows, err := cli.db.Query(query, limit, offset)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]models.Balance, 0)

	for rows.Next() {
		var addressBytes []byte
		var addressEns string
		var liquidAmount uint64
		var stakedAmount uint64
		var blockNumber uint64

		rows.Scan(&addressBytes, &liquidAmount, &stakedAmount, &blockNumber, &addressEns)
		if err = rows.Err(); err != nil {
			log.Println(err)
			return nil, err
		}

		address := common.BytesToAddress(addressBytes)

		result = append(result, models.Balance{Address: address, AddressEns: addressEns, LiquidAmount: liquidAmount, StakedAmount: stakedAmount, BlockNumber: blockNumber})
	}

	return result, nil
}

// PurgeBalanceObject purges balances less than minAmount
func (cli *DBClient) PurgeBalanceObject(minAmount uint64) error {
	query := `DELETE FROM balances WHERE liquid_amount + staked_amount < $1`

	cli.db.QueryRow(query, minAmount)

	return nil
}

// GetGlobalInt gets global int
func (cli *DBClient) GetGlobalInt(varName string) (uint64, error) {
	query := `SELECT value_int FROM globals WHERE var_name = $1`
	varInt := uint64(0)
	cli.db.QueryRow(query, varName).Scan(&varInt)
	return varInt, nil
}

// SetGlobalInt gets global int
func (cli *DBClient) SetGlobalInt(varName string, valInt uint64) error {
	sql := `
		INSERT INTO globals(var_name, value_int) VALUES ($1, $2)
		ON CONFLICT (var_name) DO UPDATE SET value_int = excluded.value_int
	`

	rows, err := cli.db.Query(sql, varName, valInt)
	if err != nil {
		log.Println(err)
		return err
	}
	rows.Close()

	return err
}

// InsertEns inserts/updates the address for a namehash
func (cli *DBClient) InsertEns(ens models.ENS) error {
	sql := `
		INSERT INTO ens(hash, address, name) VALUES (E'\\x%s', E'\\x%s', '%s')
		ON CONFLICT (hash) DO UPDATE SET address = excluded.address, name = excluded.name
	`
	adr := common.Bytes2Hex(ens.Address[:])[:]
	namehash := common.Bytes2Hex(ens.Hash[:])[:]
	rows, err := cli.db.Query(fmt.Sprintf(sql, namehash, adr, ens.Name))
	rows.Close()

	return err
}

// GetEnsName gets the ENS name for an address
func (cli *DBClient) GetEnsName(address string) (string, error) {
	query := strings.Join([]string{"SELECT name FROM ens WHERE address = E'\\\\", address[1:], "'"}, "")
	var name string
	rows := cli.db.QueryRow(query)
	err := rows.Scan(&name)
	return name, err
}

// GetEnsCount gets the count of ENS entries
func (cli *DBClient) GetEnsCount() (uint64, error) {
	query := `SELECT count(*) FROM ens`
	var count uint64
	err := cli.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// GetEnsEntries gets range of entries
func (cli *DBClient) GetEnsEntriesRange(limit uint64, offset uint64) ([]models.ENS, error) {
	if limit == 0 {
		limit = 20
	}

	query := "SELECT * FROM ens LIMIT $1 OFFSET $2"
	rows, err := cli.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.ENS

	for rows.Next() {
		var ens models.ENS
		var address, hash []byte

		rows.Scan(&hash, &address, &ens.Name)
		if err = rows.Err(); err != nil {
			return nil, err
		}

		ens.Hash.SetBytes(hash)
		ens.Address.SetBytes(address)

		result = append(result, ens)
	}

	return result, nil
}

// InsertProducer inserts/updates a producer and his stats
func (cli *DBClient) InsertProducer(producer models.Producer) error {
	sql := `
		INSERT INTO producers(address, produced_blocks_count, block_rewards) VALUES (E'\\x%s', %d, %d)
		ON CONFLICT (address) DO UPDATE
			SET produced_blocks_count = producers.produced_blocks_count + excluded.produced_blocks_count,
				block_rewards = producers.block_rewards + excluded.block_rewards
	`
	adr := common.Bytes2Hex(producer.Address[:])[:]
	blockRewards := new(big.Int).Div(producer.BlockRewards, precisionFactor).Uint64()
	rows, err := cli.db.Query(fmt.Sprintf(sql, adr, producer.ProducedBlocksCount, blockRewards))
	rows.Close()

	return err
}

// GetProducer gets the producer
func (cli *DBClient) GetProducer(address string) (*models.Producer, error) {
	query := strings.Join([]string{"SELECT * FROM producers WHERE address = E'\\\\", address[1:], "'"}, "")
	var producer models.Producer
	var producerAddress []byte
	var value uint64

	rows := cli.db.QueryRow(query)
	if err := rows.Scan(&producerAddress, &producer.ProducedBlocksCount, &value); err != nil {
		return nil, err
	}

	producer.Address.SetBytes(producerAddress)
	producer.BlockRewards = new(big.Int).Mul(new(big.Int).SetUint64(value), precisionFactor)

	return &producer, nil
}
