package main

import (
	"ebakus_server/db"
	"ebakus_server/ipc"
	"log"
	"os"

	"github.com/urfave/cli"
)

func pullNewBlocks() error {
	ipc, err := ipc.NewIPCInterface("/Users/harkal/ebakus/ebakus.ipc")
	if err != nil {
		log.Fatal("Failed to connect to ebakus", err)
	}

	number, err := ipc.GetBlockNumber()
	if err != nil {
		log.Fatal("Failed to get last block number")
	}

	log.Println(number)

	blocks, err := ipc.GetLastBlocks(1000)

	if err != nil {
		log.Fatal("Failed to get last blocks")
	}

	db, err := db.NewClient()

	if err != nil {
		log.Fatal("Failed to load db client")
	}

	err = db.InsertBlocks(blocks[:])

	if err != nil {
		log.Fatal("Failed to insert blocks")
	}

	return err
}

func main() {
	log.Print("Starting...")

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

	app.Commands = []cli.Command{
		{
			Name:    "fetchblocks",
			Aliases: []string{"f"},
			Usage:   "Fetch new blocks from ebakus node",
			Action: func(c *cli.Context) error {
				return pullNewBlocks()
			},
		},
	}

	app.Run(os.Args)
}
