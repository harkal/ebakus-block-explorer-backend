package ipc

import (
	"ebakus_server/models"
	"math/big"
	"log"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
)

type IPCInterface struct {
	cli *rpc.Client
}

func NewIPCInterface(endpoint string) (*IPCInterface, error) {
	cli, err := rpc.Dial(endpoint)
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

func (ipc *IPCInterface) GetLastBlocks(count int64) ([]*models.Block, error) {
	blockNumber, err := ipc.GetBlockNumber()
	if err != nil {
		return nil, err
	}

	first := blockNumber.Int64() 
	last := first - count
	
	blocks := make([]*models.Block, 0)
	for i := first; i > last && i >= 0 ; i-- {
		bl, err := ipc.GetBlock(big.NewInt(i))
		if err != nil {
			log.Println(err.Error())
		} else {
			blocks = append(blocks,bl)
		}
	}
	
	return blocks, nil
}