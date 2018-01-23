package main

import (
	_ "ebakus_server/web3_dao"
	"ebakus_server/web3_dao"
	// "github.com/gorilla/mux"
	"log"
	"math/big"
)

func main() {
	log.Print("Starting...")
	// TEST
	web3_dao.GetBlock(big.NewInt(2000))
}