package models

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
)

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
