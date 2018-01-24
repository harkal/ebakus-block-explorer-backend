package main

import (
	"ebakus_server/ipc"
	"ebakus_server/db"
	"log"
)

func main() {
	log.Print("Starting...")

	ipc, err := ipc.NewIPCInterface("/Users/pantelisgiazitsis/ebakus/ebakus.ipc")
	if err != nil {
		log.Fatal("Failed to connect to ebakus", err)
	}

	number, err := ipc.GetBlockNumber()
	if err != nil {
		log.Fatal("Failed to get last block number")
	}

	log.Println(number)

	block, err := ipc.GetBlock(number)
	if err != nil {
		log.Fatal("Failed to get block ", number)
	}

	log.Println(block)

	blocks,err := ipc.GetLastBlocks(1000)

	if err != nil {
		log.Fatal("Failed to get last blocks")
	}

	log.Println(blocks)

	db, err := db.NewClient()

	if err != nil {
		log.Fatal("Failed to load db client")
	}

	err = db.InsertBlocks(blocks)

	if err != nil {
		log.Fatal("Failed to insert blocks")
	}

}
