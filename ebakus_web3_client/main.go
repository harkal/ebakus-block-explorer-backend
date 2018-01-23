package main

import (
	"ebakus_server/web3_dao"
	// "github.com/gorilla/mux"
	"log"
)

func main() {
	log.Print("Starting...")

	number, err := web3_dao.GetBlockNumber()
	if err != nil {
		log.Fatal("Failed to get last block number")
	}

	log.Println(number)
}
