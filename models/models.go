package models

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
)

//go:generate gencodec -type Block -field-override blockMarshaling -out gen_block_json.go

type Block struct {
	Number           hexutil.Uint64 `json:"number"`
	TimeStamp        hexutil.Uint64 `json:"timestamp"`
	Hash             common.Hash    `json:"hash"`
	ParentHash       common.Hash    `json:"parent_hash"`
	StateRoot        common.Hash    `json:"state_root"`
	TransactionsRoot common.Hash    `json:"transactions_root"`
	ReceiptsRoot     common.Hash    `json:"receipts_root"`
	Size             hexutil.Uint64 `json:"size"`
	GasUsed          hexutil.Uint64 `json:"gas_used"`
	GasLimit         hexutil.Uint64 `json:"gas_limit"`
	Transactions     []common.Hash  `json:"-"`
	LogsBloom        types.Bloom    `json:"-"`
}

type blockMarshaling struct {
	Number    uint64
	TimeStamp uint64
	Size      uint64
	GasUsed   uint64
	GasLimit  uint64
}

type Transaction struct {
	Hash             common.Hash
	Nonce            hexutil.Uint64
	BlockHash        common.Hash
	BlockNumber      hexutil.Uint64
	TransactionIndex hexutil.Uint64
	From             common.UnprefixedAddress
	To               common.UnprefixedAddress
	value            hexutil.Uint64
	GasPrice         hexutil.Uint64
	Gas              hexutil.Uint64
	//TODO: Find type for input
	//	Input				TYPE?
}
