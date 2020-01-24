package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"bitbucket.org/pantelisss/ebakus_server/db"
	ipcModule "bitbucket.org/pantelisss/ebakus_server/ipc"
	"bitbucket.org/pantelisss/ebakus_server/models"
	"bitbucket.org/pantelisss/ebakus_server/redis"

	"github.com/nightlyone/lockfile"

	"github.com/urfave/cli/altsrc"
	cli "gopkg.in/urfave/cli.v1"
)

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

	first, err := db.GetLatestBlockNumber()
	if err != nil {
		return err
	}

	first++
	log.Printf("Going to insert blocks %d to %d (%d)", first, last, last-first+1)

	stime := time.Now()

	blockCh := make(chan *models.Block, 512)
	txsHashCh := make(chan ipcModule.TransactionWithTimestamp, 512)
	txsCh := make(chan models.TransactionFull, 512)

	var wg sync.WaitGroup

	workerThreads := c.Int("threads")
	ops := int64(workerThreads)
	for i := 0; i < workerThreads; i++ {
		wg.Add(1)
		go ipc.StreamBlocks(&wg, blockCh, txsHashCh, &ops, first, last, workerThreads, i)
	}

	wg.Add(3)
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
	app.Copyright = "(c) 2018 Ebakus Team"
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
	}

	app.Run(os.Args)
}
