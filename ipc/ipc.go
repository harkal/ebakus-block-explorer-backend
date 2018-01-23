package ipc

import (
	"ebakus_server/models"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
)

type IPCInterface struct {
	cli *rpc.Client
}

func NewIPCInterface(endpoint string) (*IPCInterface, error) {
	cli, err := rpc.Dial("/Users/harkal/ebakus/ebakus.ipc")
	if err != nil {
		return nil, err
	}

	return &IPCInterface{cli}, nil
}

//
// Get the top block number
//
func (ipc *IPCInterface) GetBlockNumber() (*big.Int, error) {
	var v hexutil.Big

	err := ipc.cli.Call(&v, "eth_blockNumber")
	if err != nil {
		return nil, err
	}

	return v.ToInt(), nil
}

func (ipc *IPCInterface) GetBlock(number *big.Int) (*models.Block, error) {
	var block models.Block

	err := ipc.cli.Call(&block, "eth_getBlockByNumber", hexutil.EncodeBig(number), false)
	if err != nil {
		return nil, err
	}

	return &block, nil
}
