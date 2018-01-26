package ipc

import (
	"ebakus_server/models"
	"errors"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
)

var (
	// ErrNoCode is returned when last is greater than first
	ErrInvalideBlockRange = errors.New("Invalid block range")
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

	blocks, err := ipc.GetBlocks(first, last)

	return blocks, err
}

func (ipc *IPCInterface) GetBlocks(first, last uint64) ([]*models.Block, error) {
	count := last - first + 1
	if count < 0 {
		return nil, ErrInvalideBlockRange
	}

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

func (ipc *IPCInterface) StreamTransactions(tCh chan *models.Transaction, hashCh chan *common.Hash) {
	for hash := range hashCh {
		if tx, err := ipc.GetTransactionByHash(hash); err == nil {
			tCh <- tx
		}
	}
	close(tCh)
}

func (ipc *IPCInterface) StreamBlocks(bCh chan *models.Block, tCh chan *common.Hash, ops *int64, first, last uint64, stride, offset int) error {
	count := last - first + 1
	if count < 0 {
		return ErrInvalideBlockRange
	}

	for i := uint64(offset); i < count; i = i + uint64(stride) {
		bl, err := ipc.GetBlock(i + first)
		if err != nil {
			return err
		}
		bCh <- bl

		for _, tx := range bl.Transactions {
			tCh <- &tx
		}
	}

	if atomic.AddInt64(ops, -1) == 0 {
		close(bCh)
		close(tCh)
	}

	return nil
}

func (ipc *IPCInterface) GetTransactionByHash(hash *common.Hash) (*models.Transaction, error) {
	var tr models.Transaction

	err := ipc.cli.Call(&tr, "eth_getTransactionByHash", hash)
	if err != nil {
		return nil, err
	}

	return &tr, nil
}
