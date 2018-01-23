package web3_dao

import (
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
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

type Block struct {
	Number           hexutil.Uint64
	TimeStamp        hexutil.Uint64
	Hash             common.Hash
	ParentHash       common.Hash
	StateRoot        common.Hash
	TransactionsRoot common.Hash
	ReceiptsRoot     common.Hash
	Size             hexutil.Uint64
	GasUsed          hexutil.Uint64
	GasLimit         hexutil.Uint64
	Transactions     []common.Hash
	LogsBloom        types.Bloom
}

func GetBlock(number *big.Int) (*Block, error) {
	var block Block

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
