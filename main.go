package main

import (
	"ebakus_server/db"
	"ebakus_server/ipc"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"github.com/urfave/cli"
)

func expandHome(path string) string {
	if path[:2] == "~/" {
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
		cli.StringFlag{
			Name:  "ipc",
			Usage: "The ebakus node to connect to e.g. ./ebakus/ebakus.ipc",
			Value: "~/ebakus/ebakus.ipc",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "fetchblocks",
			Aliases: []string{"f"},
			Usage:   "Fetch new blocks from ebakus node",
			Flags:   genericFlags,
			Action:  pullNewBlocks,
		},
	}

	app.Run(os.Args)
}
