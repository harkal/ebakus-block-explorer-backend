package main

import (
	"ebakus_server/db"
	"ebakus_server/ipc"
	"ebakus_server/models"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/urfave/cli/altsrc"
	cli "gopkg.in/urfave/cli.v1"
)

func expandHome(path string) string {
	if len(path) >= 2 && path[:2] == "~/" {
		usr, _ := user.Current()
		dir := usr.HomeDir
		path = filepath.Join(dir, path[2:])
	}
	return path
}

func pullNewBlocks(c *cli.Context) error {
	ipcFile := expandHome(c.String("ipc"))
	ipc, err := ipc.NewIPCInterface(ipcFile)
	if err != nil {
		log.Fatal("Failed to connect to ebakus", err)
	}

	dbname := c.String("dbname")
	dbhost := c.String("dbhost")
	dbport := c.Int("dbport")
	dbuser := c.String("dbuser")
	dbpass := c.String("dbpass")
	db, err := db.NewClient(dbname, dbhost, dbport, dbuser, dbpass)
	if err != nil {
		log.Fatal("Failed to load db client")
	}

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

	ch := make(chan *models.Block, 512)

	stime := time.Now()

	var ops int64 = 0

	workerThreads := 8

	for i := 0; i < workerThreads; i++ {
		go ipc.StreamBlocks(ch, &ops, first, last, workerThreads, i)
	}

	blocks := make([]*models.Block, 0)
	for block := range ch {
		if len(ch) > 500 {
			log.Println("Starving ", block.Number, len(ch))
		}
		blocks = append(blocks, block)
		if len(blocks) > 256 {
			err = db.InsertBlocks(blocks[:])
			if err != nil {
				log.Fatal("Failed to insert blocks")
			}
			blocks = make([]*models.Block, 0)
		}
	}

	err = db.InsertBlocks(blocks[:])
	if err != nil {
		log.Fatal("Failed to insert blocks")
	}

	elapsed := time.Now().Sub(stime)
	log.Printf("Processed %d blocks in %.3f (%.0f bps)", last-first+1, elapsed.Seconds(), float64(last-first+1)/elapsed.Seconds())

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
	}

	app.Run(os.Args)
}
