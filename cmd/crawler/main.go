package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"time"

	"bitbucket.org/pantelisss/ebakus_server/db"
	"bitbucket.org/pantelisss/ebakus_server/ipc"
	"bitbucket.org/pantelisss/ebakus_server/models"

	"github.com/ethereum/go-ethereum/common"

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
		// if len(ch) >= 512 {
		// 	log.Println("Chocking ", block.Number, len(ch))
		// }

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

func streamInsertTransactions(db *db.DBClient, txsCh <-chan *models.Transaction) {
	const bufSize = 400
	count := 0
	txs := make([]*models.Transaction, 0, bufSize)

	for t := range txsCh {
		if len(txsCh) >= 512 {
			log.Println("Chocking on transactions", t.Hash, len(txsCh))
		}

		if t.BlockNumber == 174950 {
			fmt.Println(">>> 174950 TX", t.Hash.String())
		}

		txs = append(txs, t)

		if len(txs) >= bufSize {
			err := db.InsertTransactions(txs[:])
			if err != nil {
				log.Println("Error streamInsertTransactions", err.Error())
			}
			count = count + len(txs)
			txs = make([]*models.Transaction, 0, bufSize)
		}
	}

	err := db.InsertTransactions(txs)
	if err != nil {
		log.Println("Error streamInsertTransactions", err.Error())
	}
	count = count + len(txs)
	fmt.Println("Finished inserting", count, "transactions")
}

func pullNewBlocks(c *cli.Context) error {
	ipcFile := expandHome(c.String("ipc"))
	ipc, err := ipc.NewIPCInterface(ipcFile)
	if err != nil {
		log.Fatal("Failed to connect to ebakus", err)
	}

	err = db.InitFromCli(c)
	if err != nil {
		log.Fatal("Failed to load db client")
	}
	db := db.GetClient()

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
	txsHashCh := make(chan *common.Hash, 512)
	txsCh := make(chan *models.Transaction, 512)

	workerThreads := c.Int("threads")
	ops := int64(workerThreads)
	for i := 0; i < workerThreads; i++ {
		go ipc.StreamBlocks(blockCh, txsHashCh, &ops, first, last, workerThreads, i)
	}

	go ipc.StreamTransactions(txsCh, txsHashCh)
	go streamInsertTransactions(db, txsCh)

	count, err := streamInsertBlocks(db, blockCh)

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
	app.Usage = "Run in various modes depending on funcion mode"

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
