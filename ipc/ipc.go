package ipc

import (
	"ebakus_server/models"
	"log"

	"github.com/ethereum/go-ethereum/common"

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
func (ipc *IPCInterface) GetBlockNumber() (uint64, error) {
	var v hexutil.Big

	err := ipc.cli.Call(&v, "eth_blockNumber")
	if err != nil {
		return 0, err
	}

	return v.ToInt().Uint64(), nil
}

func (ipc *IPCInterface) GetBlock(number uint64) (*models.Block, error) {
	var block models.Block

	err := ipc.cli.Call(&block, "eth_getBlockByNumber", hexutil.EncodeUint64(number), false)
	if err != nil {
		return nil, err
	}

	return &block, nil
}

func (ipc *IPCInterface) GetLastBlocks(count uint64) ([]*models.Block, error) {
	last, err := ipc.GetBlockNumber()
	if err != nil {
		return nil, err
	}

	first := last - count + 1
	if first < 0 {
		first = 0
	}

	log.Println(first, last, count)

	blocks := make([]*models.Block, count)
	for i := uint64(0); i < count; i++ {
		bl, err := ipc.GetBlock(i + first)
		if err != nil {
			return nil, err
		}
		blocks[i] = bl
	}

	return blocks, nil
}

func (ipc *IPCInterface) GetTransactionByHash(hash common.Hash) (*models.Transaction, error) {
	var tr models.Transaction
	err := ipc.cli.Call(&tr, "eth_getTransactionByHash", hash)
	if err != nil {
		log.Println(err.Error())

		return nil, err
	}

	log.Println(tr)

	return &tr, nil
}
