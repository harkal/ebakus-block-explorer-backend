// Code generated by github.com/fjl/gencodec. DO NOT EDIT.

package models

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
)

var _ = (*blockMarshaling)(nil)

func (b Block) MarshalJSON() ([]byte, error) {
	type Block struct {
		Number           uint64        `json:"number"`
		TimeStamp        uint64        `json:"timestamp"`
		Hash             common.Hash   `json:"hash"`
		ParentHash       common.Hash   `json:"parent_hash"`
		StateRoot        common.Hash   `json:"state_root"`
		TransactionsRoot common.Hash   `json:"transactions_root"`
		ReceiptsRoot     common.Hash   `json:"receipts_root"`
		Size             uint64        `json:"size"`
		GasUsed          uint64        `json:"gas_used"`
		GasLimit         uint64        `json:"gas_limit"`
		Transactions     []common.Hash `json:"-"`
		LogsBloom        types.Bloom   `json:"-"`
	}
	var enc Block
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
	enc.Transactions = b.Transactions
	enc.LogsBloom = b.LogsBloom
	return json.Marshal(&enc)
}

func (b *Block) UnmarshalJSON(input []byte) error {
	type Block struct {
		Number           *uint64       `json:"number"`
		TimeStamp        *uint64       `json:"timestamp"`
		Hash             *common.Hash  `json:"hash"`
		ParentHash       *common.Hash  `json:"parent_hash"`
		StateRoot        *common.Hash  `json:"state_root"`
		TransactionsRoot *common.Hash  `json:"transactions_root"`
		ReceiptsRoot     *common.Hash  `json:"receipts_root"`
		Size             *uint64       `json:"size"`
		GasUsed          *uint64       `json:"gas_used"`
		GasLimit         *uint64       `json:"gas_limit"`
		Transactions     []common.Hash `json:"-"`
		LogsBloom        *types.Bloom  `json:"-"`
	}
	var dec Block
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}
	if dec.Number != nil {
		b.Number = hexutil.Uint64(*dec.Number)
	}
	if dec.TimeStamp != nil {
		b.TimeStamp = hexutil.Uint64(*dec.TimeStamp)
	}
	if dec.Hash != nil {
		b.Hash = *dec.Hash
	}
	if dec.ParentHash != nil {
		b.ParentHash = *dec.ParentHash
	}
	if dec.StateRoot != nil {
		b.StateRoot = *dec.StateRoot
	}
	if dec.TransactionsRoot != nil {
		b.TransactionsRoot = *dec.TransactionsRoot
	}
	if dec.ReceiptsRoot != nil {
		b.ReceiptsRoot = *dec.ReceiptsRoot
	}
	if dec.Size != nil {
		b.Size = hexutil.Uint64(*dec.Size)
	}
	if dec.GasUsed != nil {
		b.GasUsed = hexutil.Uint64(*dec.GasUsed)
	}
	if dec.GasLimit != nil {
		b.GasLimit = hexutil.Uint64(*dec.GasLimit)
	}
	if dec.Transactions != nil {
		b.Transactions = dec.Transactions
	}
	if dec.LogsBloom != nil {
		b.LogsBloom = *dec.LogsBloom
	}
	return nil
}
