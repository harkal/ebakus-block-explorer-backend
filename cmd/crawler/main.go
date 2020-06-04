package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/ebakus/ebakus-block-explorer-backend/db"
	"github.com/ebakus/ebakus-block-explorer-backend/ipc"
	ipcModule "github.com/ebakus/ebakus-block-explorer-backend/ipc"
	"github.com/ebakus/ebakus-block-explorer-backend/models"
	"github.com/ebakus/ebakus-block-explorer-backend/redis"

	"github.com/ebakus/go-ebakus/common"

	"github.com/nightlyone/lockfile"

	"github.com/urfave/cli"
	"github.com/urfave/cli/altsrc"
)

const maxRichList = 1000
const maxAccountsPerRun = 1000000
const maxBlocksPerRun = 500000
const rich_list_last_block = "rich_list_last_block"

var (
	valueDecimalPoints = int64(4)
	precisionFactor    = new(big.Int).Exp(big.NewInt(10), big.NewInt(18-valueDecimalPoints), nil)
)

func doRichlist(c *cli.Context) error {
	lock, err := lockfile.New(filepath.Join(os.TempDir(), "ebakus-crawler-richlist-"+c.String("dbname")+".lock"))
	if err != nil {
		fmt.Printf("Cannot init lock. reason: %v", err)
		return err
	}
	err = lock.TryLock()
	if err != nil {
		fmt.Printf("Cannot lock %q, reason: %v", lock, err)
		return err
	}
	defer lock.Unlock()

	ipcFile := expandHome(c.String("ipc"))
	ipc, err := ipcModule.NewIPCInterface(ipcFile)
	if err != nil {
		log.Fatal("Failed to connect to ebakus", err)
	}

	err = db.InitFromCli(c)
	if err != nil {
		log.Fatal("Failed to load db client")
	}
	db := db.GetClient()

	lastBlock, err := ipc.GetBlockNumber()
	if err != nil {
		log.Fatal("Failed to get last block number")
	}

	firstBlock, err := db.GetGlobalInt(rich_list_last_block)
	if err != nil {
		log.Fatal("Failed to get first block number")
	}

	if lastBlock-firstBlock > maxBlocksPerRun {
		lastBlock = firstBlock + maxBlocksPerRun
	}

	log.Printf("Going to process blocks from %d to %d", firstBlock, lastBlock)

	accounts := make(map[common.Address]uint64)

	i := firstBlock
	for ; i < lastBlock; i++ {
		if i%50000 == 0 {
			log.Println("Fetching block: ", i)
		}

		block, err := db.GetBlockByID(i)
		if err != nil {
			break
		}

		blockNumber := uint64(block.Number)

		txs, err := db.GetTransactionsByAddress(block.Hash.Hex(), models.ADDRESS_BLOCKHASH, 0, 0xffff, "")
		if err != nil {
			break
		}

		accounts[block.Producer] = blockNumber

		for _, tx := range txs {
			accounts[tx.Tx.From] = blockNumber
			if tx.Tx.To != nil {
				accounts[*tx.Tx.To] = blockNumber
			}
		}

		// log.Println("Max accounts reached", len(accounts))
		if len(accounts) > maxAccountsPerRun {
			log.Println("Max accounts reached")
			break
		}
	}

	lastBlock = i

	// count, minLiquid, _, minStaked, _, err := db.GetBalanceStats()
	// if err != nil {
	// 	log.Println(err)
	// }
	// min := minLiquid + minStaked

	log.Println("Total accounts touched: ", len(accounts))

	// addressToBalance := make(map[common.Address]*models.Balance)

	// for _, bal := range balances {
	// 	addressToBalance[bal.Address] = &bal
	// }

	systemContractAddress := common.HexToAddress("0x0000000000000000000000000000000000000101")
	delete(accounts, systemContractAddress)

	i = 0
	for address, bn := range accounts {
		i++
		if i%50000 == 0 {
			log.Println("Updating balance: ", i, " of ", len(accounts))
		}
		// balObj := addressToBalance[address]

		bigBalance, err := ipc.GetAddressBalance(address)
		if err != nil {
			log.Println("Retrieve balance failed:", address, err)
			continue
		}
		liquid := new(big.Int).Div(bigBalance, big.NewInt(1e14)).Uint64()

		staked, err := ipc.GetAddressStaked(address)
		if err != nil {
			log.Println("Retrieve stake failed:", address, err)
			continue
		}

		// totalBalance := liquid + staked

		// if count < maxRichList {
		// 	if totalBalance < min {
		// 		continue
		// 	}
		// }

		// if balObj != nil {
		// 	balObj.Amount = totalBalance
		// } else {
		// 	addressToBalance[address] = &models.Balance{Address: address, Amount: totalBalance, BlockNumber: bn}
		// }

		db.InsertBalance(address, liquid, staked, bn)
	}

	balances, err := db.GetTopBalances(maxRichList, 0)
	if err != nil {
		log.Println(err)
	}

	if len(balances) > 0 {
		db.PurgeBalanceObject(balances[len(balances)-1].LiquidAmount + balances[len(balances)-1].StakedAmount)
	}

	err = db.SetGlobalInt(rich_list_last_block, lastBlock)
	if err != nil {
		log.Fatal("Failed to set last processed block number")
	}

	return nil
}

func getBlock(c *cli.Context) error {
	number, err := strconv.Atoi(c.Args().Get(0))
	if err != nil {
		return err
	}

	err = db.InitFromCli(c)
	if err != nil {
		return err
	}
	db := db.GetClient()

	block, err := db.GetBlockByID(uint64(number))
	if err != nil {
		return err
	}

	jsonBlock := models.JSONBlock(*block)

	json, _ := json.MarshalIndent(jsonBlock, "", "  ")
	fmt.Printf("%s\n", json)

	return nil
}

func expandHome(path string) string {
	if len(path) >= 2 && path[:2] == "~/" {
		usr, _ := user.Current()
		dir := usr.HomeDir
		path = filepath.Join(dir, path[2:])
	}
	return path
}

func streamInsertBlocks(db *db.DBClient, ch chan *models.Block) (int, error) {
	const bufSize = 400
	count := 0
	blocks := make([]*models.Block, 0, bufSize)

	for block := range ch {
		if len(ch) >= 512 {
			log.Println("Chocking ", block.Number, len(ch))
		}

		blocks = append(blocks, block)

		if len(blocks) >= bufSize {
			err := db.InsertBlocks(blocks[:])
			if err != nil {
				return 0, err
			}
			count = count + len(blocks)
			blocks = make([]*models.Block, 0, bufSize)
		}
	}

	err := db.InsertBlocks(blocks[:])
	if err != nil {
		return 0, err
	}

	count = count + len(blocks)

	return count, nil
}

func streamInsertTransactions(wg *sync.WaitGroup, db *db.DBClient, txsCh <-chan models.TransactionFull) {
	defer wg.Done()
	const bufSize = 20
	count := 0
	txs := make([]models.TransactionFull, 0, bufSize)

	for t := range txsCh {
		if len(txsCh) >= 512 {
			log.Println("Chocking on transactions", t.Tx.Hash, len(txsCh))
		}

		txs = append(txs, t)

		if len(txs) >= bufSize {
			err := db.InsertTransactions(txs[:])
			if err != nil {
				log.Println("Error streamInsertTransactions", err.Error())
			}
			count = count + len(txs)
			txs = make([]models.TransactionFull, 0, bufSize)
		}
	}

	err := db.InsertTransactions(txs[:])
	if err != nil {
		log.Println("Error streamInsertTransactions", err.Error())
	}
	count = count + len(txs)
	fmt.Println("Finished inserting", count, "transactions")
}

func streamDeleteBlockWithTransactions(wg *sync.WaitGroup, db *db.DBClient, dCh <-chan *models.Block, bCh chan<- *models.Block, tCh chan<- ipc.TransactionWithTimestamp, pCh chan<- common.Address) {
	defer wg.Done()

	for bl := range dCh {
		err := db.DeleteBlockWithTransactionsByID(uint64(bl.Number), bl.Producer)

		if err != nil {
			log.Println("Error streamDeleteBlockWithTransactions", err.Error())
			continue
			// TODO: exit here?

		}

		bCh <- bl
		pCh <- bl.Producer

		for _, tx := range bl.Transactions {
			tCh <- ipc.TransactionWithTimestamp{Hash: tx, Timestamp: bl.TimeStamp}
		}
	}

	close(bCh)
	close(tCh)
	close(pCh)
}

func streamInsertProducers(wg *sync.WaitGroup, db *db.DBClient, pCh chan common.Address) (int, error) {
	defer wg.Done()

	const bufSize = 50
	count := 0
	producers := make(map[common.Address]*models.Producer)

	blockRewardsWei := new(big.Int).Mul(new(big.Int).SetUint64(3171), precisionFactor)

	for address := range pCh {
		if _, ok := producers[address]; !ok {
			producers[address] = &models.Producer{
				Address:             address,
				ProducedBlocksCount: 1,
				BlockRewards:        blockRewardsWei,
			}
		} else {
			producers[address].ProducedBlocksCount++
			producers[address].BlockRewards = new(big.Int).Add(producers[address].BlockRewards, blockRewardsWei)
		}

		if len(producers) >= bufSize {
			for _, prod := range producers {
				err := db.InsertProducer(*prod)
				if err != nil {
					return 0, err
				}
				count++
				delete(producers, prod.Address)
			}
		}
	}

	for _, prod := range producers {
		err := db.InsertProducer(*prod)
		if err != nil {
			return 0, err
		}
		count++
		delete(producers, prod.Address)
	}

	return count, nil
}

func pullNewBlocks(c *cli.Context) error {
	lock, err := lockfile.New(filepath.Join(os.TempDir(), "ebakus-crawler-"+c.String("dbname")+".lock"))
	if err != nil {
		fmt.Printf("Cannot init lock. reason: %v", err)
		return err
	}
	err = lock.TryLock()
	if err != nil {
		fmt.Printf("Cannot lock %q, reason: %v", lock, err)
		return err
	}
	defer lock.Unlock()

	ipcFile := expandHome(c.String("ipc"))
	ipc, err := ipcModule.NewIPCInterface(ipcFile)
	if err != nil {
		log.Fatal("Failed to connect to ebakus", err)
	}

	err = db.InitFromCli(c)
	if err != nil {
		log.Fatal("Failed to load db client")
	}
	db := db.GetClient()

	if err := redis.InitFromCli(c); err != nil {
		log.Fatal("Failed to connect to redis", err)
	}
	defer redis.Pool.Close()

	last, err := ipc.GetBlockNumber()
	if err != nil {
		log.Fatal("Failed to get last block number")
	}

	log.Printf("Going to insert blocks backwards from %d", last)

	stime := time.Now()

	deleteCh := make(chan *models.Block, 512)
	blockCh := make(chan *models.Block, 512)
	txsHashCh := make(chan ipcModule.TransactionWithTimestamp, 512)
	txsCh := make(chan models.TransactionFull, 512)
	producerCh := make(chan common.Address, 512)

	var wg sync.WaitGroup

	wg.Add(6)
	go ipc.StreamBlocks(&wg, db, blockCh, txsHashCh, producerCh, deleteCh, last)
	go streamInsertProducers(&wg, db, producerCh)
	go streamDeleteBlockWithTransactions(&wg, db, deleteCh, blockCh, txsHashCh, producerCh)

	go ipc.StreamTransactions(&wg, db, txsCh, txsHashCh)
	go streamInsertTransactions(&wg, db, txsCh)

	var count int
	go func() {
		defer wg.Done()
		count, _ = streamInsertBlocks(db, blockCh)
	}()

	wg.Wait()

	elapsed := time.Now().Sub(stime)
	log.Printf("Processed %d blocks in %.3f (%.0f bps)", count, elapsed.Seconds(), float64(count)/elapsed.Seconds())

	return err
}

func doEnsSync(c *cli.Context) error {
	lock, err := lockfile.New(filepath.Join(os.TempDir(), "ebakus-crawler-enssync-"+c.String("dbname")+".lock"))
	if err != nil {
		fmt.Printf("Cannot init lock. reason: %v", err)
		return err
	}
	err = lock.TryLock()
	if err != nil {
		fmt.Printf("Cannot lock %q, reason: %v", lock, err)
		return err
	}
	defer lock.Unlock()

	ipcFile := expandHome(c.String("ipc"))
	ipc, err := ipcModule.NewIPCInterface(ipcFile)
	if err != nil {
		log.Fatal("Failed to connect to ebakus", err)
	}

	err = db.InitFromCli(c)
	if err != nil {
		log.Fatal("Failed to load db client")
	}
	db := db.GetClient()

	ensContractAddress := common.HexToAddress(c.String("enscontractaddress"))
	zeroAddress := common.Address{}
	if ensContractAddress == zeroAddress {
		log.Fatal("No contract address defined for the ENS contract")
	}

	log.Printf("Going to sync up ENS names with its addresses")

	stime := time.Now()

	numberOfEntries, err := db.GetEnsCount()
	if err != nil {
		log.Fatal("Failed to get number of ENS entries in DB", err.Error())
	}
	if numberOfEntries == 0 {
		log.Println("No ENS entries to process")
		return nil
	}

	updatedEntries := 0
	const chunkSize = 100

	for i := uint64(0); i < numberOfEntries; i += chunkSize {
		entries, err := db.GetEnsEntriesRange(chunkSize, i)
		if err != nil {
			log.Fatal("Failed to get ENS entries from DB", err.Error())
			return err
		}

		for _, ens := range entries {
			addr, err := ipc.GetENSAddress(ensContractAddress, ens.Hash)
			if err != nil {
				log.Println("Failed to get ENS address from node state:", err.Error())
				continue
			}

			if ens.Address == addr {
				continue
			}

			ens.Address = addr
			updatedEntries++

			err = db.InsertEns(ens)
			if err != nil {
				log.Fatal("Error InsertEns failed", err.Error())
			}
		}
	}

	elapsed := time.Now().Sub(stime)
	log.Printf("Updated %d of %d names in %.3f (%.0f bps)", updatedEntries, numberOfEntries, elapsed.Seconds(), float64(numberOfEntries)/elapsed.Seconds())

	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "Ebakus Blockchain Explorer"
	app.Version = "0.0.1"
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "Harry Kalogirou",
			Email: "harkal@gmail.com",
		},
		cli.Author{
			Name:  "Pantelis Giazitsis",
			Email: "burn665@gmail.com",
		},
	}
	app.Copyright = "(c) 2020 Ebakus Team"
	app.Usage = "Run in various modes depending on function mode"

	genericFlags := []cli.Flag{
		altsrc.NewStringFlag(cli.StringFlag{
			Name:  "ipc",
			Usage: "The ebakus node to connect to e.g. ./ebakus/ebakus.ipc",
			Value: "~/ebakus/ebakus.ipc",
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name:  "dbhost",
			Value: "localhost",
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name:  "dbname",
			Value: "ebakus",
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name:  "dbuser",
			Value: "ebakus",
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name:  "dbpass",
			Value: "",
		}),
		altsrc.NewIntFlag(cli.IntFlag{
			Name:  "dbport",
			Value: 5432,
		}),
		altsrc.NewIntFlag(cli.IntFlag{
			Name:  "threads",
			Value: 8,
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name:  "redishost",
			Value: "localhost",
		}),
		altsrc.NewIntFlag(cli.IntFlag{
			Name:  "redisport",
			Value: 6379,
		}),
		altsrc.NewIntFlag(cli.IntFlag{
			Name:  "redispoolsize",
			Value: 10,
		}),
		altsrc.NewIntFlag(cli.IntFlag{
			Name:  "redisdbselect",
			Value: 0,
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name:  "enscontractaddress",
			Value: "",
		}),
		cli.StringFlag{
			Name:  "config",
			Value: "config.yaml",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "fetchblocks",
			Aliases: []string{"f"},
			Usage:   "Fetch new blocks from ebakus node",
			Before:  altsrc.InitInputSourceWithContext(genericFlags, altsrc.NewYamlSourceFromFlagFunc("config")),
			Flags:   genericFlags,
			Action:  pullNewBlocks,
		},
		{
			Name:    "getblock",
			Aliases: []string{"gb"},
			Usage:   "Retrieve block from database",
			Before:  altsrc.InitInputSourceWithContext(genericFlags, altsrc.NewYamlSourceFromFlagFunc("config")),
			Flags:   genericFlags,
			Action:  getBlock,
		},
		{
			Name:    "computerich",
			Aliases: []string{"cr"},
			Usage:   "Compute richlist",
			Before:  altsrc.InitInputSourceWithContext(genericFlags, altsrc.NewYamlSourceFromFlagFunc("config")),
			Flags:   genericFlags,
			Action:  doRichlist,
		},
		{
			Name:    "enssync",
			Aliases: []string{"ens"},
			Usage:   "ENS names will sync up its address",
			Before:  altsrc.InitInputSourceWithContext(genericFlags, altsrc.NewYamlSourceFromFlagFunc("config")),
			Flags:   genericFlags,
			Action:  doEnsSync,
		},
	}

	app.Run(os.Args)
}
