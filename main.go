package main

import (
	"ebakus_server/ipc"
	"log"
)

func main() {
	log.Print("Starting...")

	ipc, err := ipc.NewIPCInterface("~/ebakus/ebakus.ipc")
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
}
