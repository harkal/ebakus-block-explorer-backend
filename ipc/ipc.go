package web3_dao

import (
	"ebakus_server/models"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
)

var cli *rpc.Client

func init() {
	fmt.Println("init()")

	var err error
	/*** Test with rpc ***/
	cli, err = rpc.Dial("/Users/harkal/ebakus/ebakus.ipc")
	if err != nil {
		log.Fatal("Failed to Dial", err.Error())
	}
}

//
// Get the top block number
//
func GetBlockNumber() (*big.Int, error) {
	var v hexutil.Big

	err := cli.Call(&v, "eth_blockNumber")
	if err != nil {
		return nil, err
	}

	return v.ToInt(), nil
}

func GetBlock(number *big.Int) (*models.Block, error) {
	var block models.Block

	err := cli.Call(&block, "eth_getBlockByNumber", hexutil.EncodeBig(number), false)
	if err != nil {
		return nil, err
	}

	return &block, nil
}

// func SyncDatabase() {
// 	blockNumber, err := GetBlockNumber()
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}

// }
