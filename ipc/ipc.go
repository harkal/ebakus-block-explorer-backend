package ipc

import (
	"errors"
	"log"
	"sync"
	"sync/atomic"

	"bitbucket.org/pantelisss/ebakus_server/db"
	"bitbucket.org/pantelisss/ebakus_server/models"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
)

var (
	// ErrNoCode is returned when last is greater than first
	ErrInvalideBlockRange = errors.New("Invalid block range")
)

type TransactionWithTimestamp struct {
	Hash      common.Hash
	Timestamp hexutil.Uint64
}

type IPCInterface struct {
	cli *rpc.Client
}

var ipci *IPCInterface

func NewIPCInterface(endpoint string) (*IPCInterface, error) {
	cli, err := rpc.Dial(endpoint)
	if err != nil {
		return nil, err
	}

	ipci = &IPCInterface{cli}

	return ipci, nil
}

// GetIPC returns the current ipc instance.
// Dev Commentary: I'm sorry for this but I needed a way to have
// the IPC available throughout the project. If you know
// a better way to do this I'd like to know it too.
func GetIPC() *IPCInterface {
	return ipci
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

func (ipc *IPCInterface) StreamTransactions(wg *sync.WaitGroup, db *db.DBClient, tCh chan<- models.TransactionFull, hashCh <-chan TransactionWithTimestamp) {
	defer wg.Done()
	for obj := range hashCh {
		tx, txr, err := ipc.GetTransactionByHash(&obj.Hash)
		if err != nil {
			log.Println("Error getTransaction ipc:", err, "hash", obj.Hash.Hex())
			continue
		}

		tx.Timestamp = obj.Timestamp

		tf := models.TransactionFull{Tx: tx, Txr: txr}
		tCh <- tf
	}
	close(tCh)
}

func (ipc *IPCInterface) StreamBlocks(wg *sync.WaitGroup, bCh chan<- *models.Block, tCh chan<- TransactionWithTimestamp, ops *int64, first, last uint64, stride, offset int) error {
	defer wg.Done()
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
			tCh <- TransactionWithTimestamp{Hash: tx, Timestamp: bl.TimeStamp}
		}
	}

	if atomic.AddInt64(ops, -1) == 0 {
		close(bCh)
		close(tCh)
	}

	return nil
}

func (ipc *IPCInterface) GetTransactionByHash(hash *common.Hash) (*models.Transaction, *models.TransactionReceipt, error) {
	var tx models.Transaction
	var txr models.TransactionReceipt

	err := ipc.cli.Call(&tx, "eth_getTransactionByHash", hash.String())
	if err != nil {
		return nil, nil, err
	}

	err = ipc.cli.Call(&txr, "eth_getTransactionReceipt", hash.String())
	if err != nil {
		return nil, nil, err
	}

	return &tx, &txr, nil
}

func (ipc *IPCInterface) GetDelegates(number uint64) ([]models.DelegateVoteInfo, error) {
	var di []models.DelegateVoteInfo

	err := ipc.cli.Call(&di, "dpos_getDelegates", hexutil.EncodeUint64(number))
	if err != nil {
		return nil, err
	}

	return di, nil
}

func (ipc *IPCInterface) GetABIForContract(address common.Address) (string, error) {
	var abi string

	err := ipc.cli.Call(&abi, "eth_getABIForContract", address)
	if err != nil {
		return "", err
	}

	return abi, nil
}
