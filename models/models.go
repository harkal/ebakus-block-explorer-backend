package models

import (
	"encoding/json"

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

type JSONBlock Block

// MarshalJSON converts a Block to a byte array
// that contains it's data in JSON format.
func (b Block) MarshalJSON() ([]byte, error) {
	// Struct with only the fields we want in the final JSON?
	var enc struct {
		Number           uint64      `json:"number"`
		TimeStamp        uint64      `json:"timestamp"`
		Hash             common.Hash `json:"hash"`
		ParentHash       common.Hash `json:"parentHash"`
		StateRoot        common.Hash `json:"stateRoot"`
		TransactionsRoot common.Hash `json:"transactionsRoot"`
		ReceiptsRoot     common.Hash `json:"receiptsRoot"`
		Size             uint64      `json:"size"`
		GasUsed          uint64      `json:"gasUsed"`
		GasLimit         uint64      `json:"gasLimit"`
	}

	enc.Number = uint64(b.Number)
	enc.TimeStamp = uint64(b.TimeStamp)
	enc.Hash = b.Hash
	enc.ParentHash = b.ParentHash
	enc.StateRoot = b.StateRoot
	enc.TransactionsRoot = b.TransactionsRoot
	enc.ReceiptsRoot = b.ReceiptsRoot
	enc.Size = uint64(b.Size)
	enc.GasUsed = uint64(b.GasUsed)
	enc.GasLimit = uint64(b.GasLimit)

	return json.Marshal(&enc)
}

type Transaction struct {
	Hash             common.Hash              `json:"hash"`
	Nonce            hexutil.Uint64           `json:"nonce"`
	BlockHash        common.Hash              `json:"blockHash"`
	BlockNumber      hexutil.Uint64           `json:"blockNumber"`
	TransactionIndex hexutil.Uint64           `json:"txIndex"`
	From             common.UnprefixedAddress `json:"from"`
	To               common.UnprefixedAddress `json:"to"`
	Value            hexutil.Uint64           `json:"value"`
	GasPrice         hexutil.Uint64           `json:"gasPrice"`
	Gas              hexutil.Uint64           `json:"gas"`
	//TODO: Find type for input
	//	Input				TYPE?
}

// MarshalJSON converts a Transaction to a byte array
// that contains it's data in JSON format.
func (b Transaction) MarshalJSON() ([]byte, error) {
	// Transactions don't need a subset of the fields
	// currently, so we don't have to follow a similar
	// procedure like the Block marshalling.
	return json.Marshal(&b)
}
