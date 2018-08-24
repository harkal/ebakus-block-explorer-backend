package models

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
)

//go:generate gencodec -type Block -field-override blockMarshaling -out gen_block_json.go

type Block struct {
	Number           hexutil.Uint64   `json:"number"`
	TimeStamp        hexutil.Uint64   `json:"timestamp"`
	Hash             common.Hash      `json:"hash"`
	ParentHash       common.Hash      `json:"parentHash"`
	TransactionsRoot common.Hash      `json:"transactionsRoot"`
	ReceiptsRoot     common.Hash      `json:"receiptsRoot"`
	Size             hexutil.Uint64   `json:"size"`
	TransactionCount hexutil.Uint64   `json:"transactionCount"`
	GasUsed          hexutil.Uint64   `json:"gasUsed"`
	GasLimit         hexutil.Uint64   `json:"gasLimit"`
	Transactions     []common.Hash    `json:"transactions"`
	Delegates        []common.Address `json:"delegates"`
	LogsBloom        types.Bloom      `json:"logBloom"`
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
		TransactionsRoot common.Hash `json:"transactionsRoot"`
		ReceiptsRoot     common.Hash `json:"receiptsRoot"`
		Size             uint64      `json:"size"`
		TransactionCount uint64      `json:"transactionCount"`
		GasUsed          uint64      `json:"gasUsed"`
		GasLimit         uint64      `json:"gasLimit"`
	}

	enc.Number = uint64(b.Number)
	enc.TimeStamp = uint64(b.TimeStamp)
	enc.Hash = b.Hash
	enc.ParentHash = b.ParentHash
	enc.TransactionsRoot = b.TransactionsRoot
	enc.ReceiptsRoot = b.ReceiptsRoot
	enc.Size = uint64(b.Size)
	enc.TransactionCount = uint64(b.TransactionCount)
	enc.GasUsed = uint64(b.GasUsed)
	enc.GasLimit = uint64(b.GasLimit)

	return json.Marshal(&enc)
}

type Transaction struct {
	Hash             common.Hash    `json:"hash"`
	Nonce            hexutil.Uint64 `json:"nonce"`
	BlockHash        common.Hash    `json:"blockHash"`
	BlockNumber      hexutil.Uint64 `json:"blockNumber"`
	TransactionIndex hexutil.Uint64 `json:"transactionIndex"`
	From             common.Address `json:"from"`
	To               common.Address `json:"to"`
	Value            hexutil.Uint64 `json:"value"`
	GasLimit         hexutil.Uint64 `json:"gas"`
	// Input            []byte         `json:"input"` // Causes error during JSON unmarshaling
}

type TransactionReceipt struct {
	Status            hexutil.Uint64 `json:"status"`
	GasUsed           hexutil.Uint64 `json:"gasUsed"`
	CumulativeGasUsed hexutil.Uint64 `json:"cumulativeGasUsed"`
}

type TransactionFull struct {
	Tx  *Transaction
	Txr *TransactionReceipt
}

type AddressType int

const (
	ADDRESS_FROM AddressType = iota
	ADDRESS_TO
	ADDRESS_BLOCKHASH
)

// MarshalJSON converts a Transaction to a byte array
// that contains it's data in JSON format.
func (t Transaction) MarshalJSON() ([]byte, error) {
	// Struct with only the fields we want in the final JSON?
	var enc struct {
		Hash             common.Hash    `json:"hash"`
		Nonce            uint64         `json:"nonce"`
		BlockHash        common.Hash    `json:"blockHash"`
		BlockNumber      uint64         `json:"blockNumber"`
		TransactionIndex uint64         `json:"transactionIndex"`
		From             common.Address `json:"from"`
		To               common.Address `json:"to"`
		Value            uint64         `json:"value"`
		GasLimit         uint64         `json:"gas"`
	}

	enc.Hash = t.Hash
	enc.Nonce = uint64(t.Nonce)
	enc.BlockHash = t.BlockHash
	enc.BlockNumber = uint64(t.BlockNumber)
	enc.TransactionIndex = uint64(t.TransactionIndex)
	enc.From = t.From
	enc.To = t.To
	enc.Value = uint64(t.Value)
	enc.GasLimit = uint64(t.GasLimit)

	return json.Marshal(&enc)
}
