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
	ParentHash       common.Hash    `json:"parentHash"`
	StateRoot        common.Hash    `json:"stateRoot"`
	TransactionsRoot common.Hash    `json:"transactionsRoot"`
	ReceiptsRoot     common.Hash    `json:"receiptsRoot"`
	Size             hexutil.Uint64 `json:"size"`
	GasUsed          hexutil.Uint64 `json:"gasUsed"`
	GasLimit         hexutil.Uint64 `json:"gasLimit"`
	Transactions     []common.Hash  `json:"transactions"`
	LogsBloom        types.Bloom    `json:"logBloom"`
}

type blockMarshaling struct {
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
