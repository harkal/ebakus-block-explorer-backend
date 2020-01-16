package models

import (
	"encoding/hex"
	"encoding/json"
	"math/big"

	"github.com/ebakus/go-ebakus/common"
	"github.com/ebakus/go-ebakus/common/hexutil"
)

//go:generate gencodec -type Block -field-override blockMarshaling -out gen_block_json.go

type SignatureType []byte

type Block struct {
	Number           hexutil.Uint64   `json:"number"`
	TimeStamp        hexutil.Uint64   `json:"timestamp"`
	Hash             common.Hash      `json:"hash"`
	ParentHash       common.Hash      `json:"parentHash"`
	Signature        SignatureType    `json:"signature"`
	TransactionsRoot common.Hash      `json:"transactionsRoot"`
	ReceiptsRoot     common.Hash      `json:"receiptsRoot"`
	Size             hexutil.Uint64   `json:"size"`
	TransactionCount hexutil.Uint64   `json:"transactionCount"`
	GasUsed          hexutil.Uint64   `json:"gasUsed"`
	GasLimit         hexutil.Uint64   `json:"gasLimit"`
	Transactions     []common.Hash    `json:"transactions"`
	Delegates        []common.Address `json:"delegates"`
	Producer         common.Address   `json:"producer"`
}

type JSONBlock Block

// MarshalJSON converts a Block to a byte array
// that contains it's data in JSON format.
func (b Block) MarshalJSON() ([]byte, error) {
	// Struct with only the fields we want in the final JSON?
	var enc struct {
		Number           uint64           `json:"number"`
		TimeStamp        uint64           `json:"timestamp"`
		Hash             common.Hash      `json:"hash"`
		ParentHash       common.Hash      `json:"parentHash"`
		Signature        SignatureType    `json:"signature"`
		TransactionsRoot common.Hash      `json:"transactionsRoot"`
		ReceiptsRoot     common.Hash      `json:"receiptsRoot"`
		Size             uint64           `json:"size"`
		TransactionCount uint64           `json:"transactionCount"`
		GasUsed          uint64           `json:"gasUsed"`
		GasLimit         uint64           `json:"gasLimit"`
		Delegates        []common.Address `json:"delegates"`
		Producer         common.Address   `json:"producer"`
	}

	enc.Number = uint64(b.Number)
	enc.TimeStamp = uint64(b.TimeStamp)
	enc.Hash = b.Hash
	enc.ParentHash = b.ParentHash
	enc.Signature = b.Signature
	enc.TransactionsRoot = b.TransactionsRoot
	enc.ReceiptsRoot = b.ReceiptsRoot
	enc.Size = uint64(b.Size)
	enc.TransactionCount = uint64(b.TransactionCount)
	enc.GasUsed = uint64(b.GasUsed)
	enc.GasLimit = uint64(b.GasLimit)
	enc.Delegates = b.Delegates
	enc.Producer = b.Producer

	return json.Marshal(&enc)
}

type InputData []byte

type Transaction struct {
	Hash             common.Hash     `json:"hash"`
	Timestamp        hexutil.Uint64  `json:"timestamp"`
	Nonce            hexutil.Uint64  `json:"nonce"`
	BlockHash        common.Hash     `json:"blockHash"`
	BlockNumber      hexutil.Uint64  `json:"blockNumber"`
	TransactionIndex hexutil.Uint64  `json:"transactionIndex"`
	From             common.Address  `json:"from"`
	To               *common.Address `json:"to"`
	Value            hexutil.Big     `json:"value"`
	GasLimit         hexutil.Uint64  `json:"gas"`
	GasPrice         hexutil.Uint64  `json:"gasPrice"`
	WorkNonce        hexutil.Uint64  `json:"workNonce"`
	Input            InputData       `json:"input"`
}

type TransactionReceipt struct {
	Status            hexutil.Uint64  `json:"status"`
	GasUsed           hexutil.Uint64  `json:"gasUsed"`
	CumulativeGasUsed hexutil.Uint64  `json:"cumulativeGasUsed"`
	ContractAddress   *common.Address `json:"contractAddress"`
}

type TransactionFull struct {
	Tx  *Transaction
	Txr *TransactionReceipt
}

type AddressType int

const (
	ADDRESS_FROM AddressType = iota
	ADDRESS_TO
	ADDRESS_ALL
	ADDRESS_BLOCKHASH
	LATEST
)

type TimestampCondition int

const (
	TIMESTAMP_EQUAL TimestampCondition = iota
	TIMESTAMP_SMALLER_EQUAL_THAN
	TIMESTAMP_GREATER_EQUAL_THAN
)

// MarshalJSON converts a Transaction to a byte array
// that contains it's data in JSON format.
func (t Transaction) MarshalJSON() ([]byte, error) {
	// Struct with only the fields we want in the final JSON?
	var enc struct {
		Hash             common.Hash     `json:"hash"`
		Timestamp        uint64          `json:"timestamp"`
		Nonce            uint64          `json:"nonce"`
		BlockHash        common.Hash     `json:"blockHash"`
		BlockNumber      uint64          `json:"blockNumber"`
		TransactionIndex uint64          `json:"transactionIndex"`
		From             common.Address  `json:"from"`
		To               *common.Address `json:"to"`
		Value            *big.Int        `json:"value"`
		GasLimit         uint64          `json:"gas"`
		GasPrice         uint64          `json:"gasPrice"`
		WorkNonce        uint64          `json:"workNonce"`
		Input            string          `json:"input"`
	}

	enc.Hash = t.Hash
	enc.Timestamp = uint64(t.Timestamp)
	enc.Nonce = uint64(t.Nonce)
	enc.BlockHash = t.BlockHash
	enc.BlockNumber = uint64(t.BlockNumber)
	enc.TransactionIndex = uint64(t.TransactionIndex)
	enc.From = t.From
	enc.To = t.To
	enc.Value = t.Value.ToInt()
	enc.GasLimit = uint64(t.GasLimit)
	enc.GasPrice = uint64(t.GasPrice)
	enc.WorkNonce = uint64(t.WorkNonce)

	enc.Input = "0x" + hex.EncodeToString(t.Input)

	return json.Marshal(&enc)
}

func (t *InputData) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err != nil {
		return err
	}

	if str == "0x" {
		return nil
	}

	*t, err = hex.DecodeString(str[2:])
	if err != nil {
		return err
	}

	return nil
}

// MarshalJSON converts a Transaction to a byte array
// that contains it's data in JSON format.
func (tf TransactionFull) MarshalJSON() ([]byte, error) {
	// Struct with only the fields we want in the final JSON?
	var enc struct {
		Hash              common.Hash     `json:"hash"`
		Timestamp         uint64          `json:"timestamp"`
		Status            uint64          `json:"status"`
		Nonce             uint64          `json:"nonce"`
		BlockHash         common.Hash     `json:"blockHash"`
		BlockNumber       uint64          `json:"blockNumber"`
		TransactionIndex  uint64          `json:"transactionIndex"`
		From              common.Address  `json:"from"`
		To                *common.Address `json:"to"`
		Value             *big.Int        `json:"value"`
		GasUsed           uint64          `json:"gasUsed"`
		CumulativeGasUsed uint64          `json:"cumulativeGasUsed"`
		GasLimit          uint64          `json:"gasLimit"`
		GasPrice          uint64          `json:"gasPrice"`
		WorkNonce         uint64          `json:"workNonce"`
		ContractAddress   *common.Address `json:"contractAddress"`
		Input             string          `json:"input"`
	}

	t := tf.Tx
	r := tf.Txr

	enc.Hash = t.Hash
	enc.Timestamp = uint64(t.Timestamp)
	enc.Status = uint64(r.Status)
	enc.Nonce = uint64(t.Nonce)
	enc.BlockHash = t.BlockHash
	enc.BlockNumber = uint64(t.BlockNumber)
	enc.TransactionIndex = uint64(t.TransactionIndex)
	enc.From = t.From
	enc.To = t.To
	enc.Value = t.Value.ToInt()
	enc.GasUsed = uint64(r.GasUsed)
	enc.CumulativeGasUsed = uint64(r.CumulativeGasUsed)
	enc.GasLimit = uint64(t.GasLimit)
	enc.GasPrice = uint64(t.GasPrice)
	enc.WorkNonce = uint64(t.WorkNonce)
	enc.ContractAddress = r.ContractAddress
	enc.Input = "0x" + hex.EncodeToString(t.Input)

	return json.Marshal(&enc)
}

type AddressResult struct {
	Address      common.Address `json:"address"`
	Balance      *big.Int       `json:"balance"`
	Stake        uint64         `json:"stake"`
	TxCount      uint64         `json:"tx_count"`
	BlockRewards *big.Int       `json:"block_rewards"`
}

type DelegateInfo struct {
	Address         common.Address `json:"address"`
	SecondsExamined uint64         `json:"seconds_examined"`
	MissedBlocks    uint64         `json:"missed_blocks"`
	TotalBlocks     uint64         `json:"total_blocks"`
	Density         float64        `json:"density"`
	Stake           uint64         `json:"stake"`
}

type DelegateVoteInfo struct {
	Address common.Address `json:"address"`
	Stake   uint64         `json:"stake"`
}
