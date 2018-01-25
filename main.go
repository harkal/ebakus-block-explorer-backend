package main

import (
	"ebakus_server/db"
	"ebakus_server/ipc"
	"log"
	"os"
	"os/user"
	"path/filepath"

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

	blocks, err := ipc.GetBlocks(first, last)
	if err != nil {
		log.Fatal("Failed to get blocks")
	}

	log.Printf("Going to insert blocks %d to %d (%d)", first, last, len(blocks))

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
