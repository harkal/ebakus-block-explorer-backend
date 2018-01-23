package main

import (
	"ebakus_server/web3_dao"
	"log"
)

func main() {
	log.Print("Starting...")

	number, err := web3_dao.GetBlockNumber()
	if err != nil {
		log.Fatal("Failed to get last block number")
	}

	log.Println(number)

	block, err := web3_dao.GetBlock(number)
	if err != nil {
		log.Fatal("Failed to get last block number")
	}

	log.Println(block)
}
