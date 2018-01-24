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

type Transaction struct {
	Hash				common.Hash
	Nonce				hexutil.Uint64
	BlockHash 			common.Hash
	BlockNumber 		hexutil.Uint64
	TransactionIndex 	hexutil.Uint64
	From				common.UnprefixedAddress
	To					common.UnprefixedAddress
	value 				hexutil.Uint64
	GasPrice			hexutil.Uint64
	Gas					hexutil.Uint64
	//TODO: Find type for input
//	Input				TYPE?
}