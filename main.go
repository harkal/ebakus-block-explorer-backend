package main

import (
	"ebakus_server/db"
	"ebakus_server/ipc"
	"log"
)

func main() {
	log.Print("Starting...")

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

}
