package models

import (
	"encoding/json"
	"log"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type Block struct {
	number int64 `json:"number"`
	size int64 `json:"size"`
	hash string `json:"hash"`
	ts int64 `json:"ts"`
	gasUsed int64 `json:"gasUsed"`
}

// Web3 keys
const  (
	numberKey = "number"
	sizeKey = "size"
	hashKey = "hash"
	tsKey = "timestamp"
	gasUsedKey = "gasUsed"
)

func NewBlockFromWeb3Map(m map[string]*json.RawMessage) *Block {
	block := new(Block)

	// number
	var numStr hexutil.Big
	err := json.Unmarshal(*m[numberKey], &numStr)
	if err != nil {
		log.Print(err.Error())
		block.number = 0 
	} else {
		block.number = numStr.ToInt().Int64()
	}

	// size
	var sizeStr hexutil.Big
	err = json.Unmarshal(*m[sizeKey], &sizeStr)
	if err != nil {
		log.Print(err.Error())
		block.size = 0 
	} else {
		block.size = sizeStr.ToInt().Int64()
	}
	
	// hash
	var hashStr string
	err = json.Unmarshal(*m[hashKey], &hashStr)
	if err != nil {
		log.Print(err.Error())
		block.hash = ""
	} else {
		block.hash = hashStr
	}

	// ts
	var tsStr hexutil.Big
	err = json.Unmarshal(*m[tsKey], &tsStr)
	if err != nil {
		log.Print(err.Error())
		block.ts = 0 
	} else {
		block.ts = tsStr.ToInt().Int64()
	}
	
	// gasUsed
	var gasUsedStr hexutil.Big
	err = json.Unmarshal(*m[gasUsedKey], &gasUsedStr)
	if err != nil {
		log.Print(err.Error())
		block.gasUsed = 0 
	} else {
		block.gasUsed = gasUsedStr.ToInt().Int64()
	}

	return block 
}