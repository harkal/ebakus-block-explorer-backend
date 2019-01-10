package webapi

import (
	"errors"

	"bitbucket.org/pantelisss/ebakus_server/db"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

var (
	blockDensityLookBackTime = 360 // seconds
)

// DPOSConfig values on running node
// TODO: get those dynamically
var (
	DPOSConfigPeriod         = 1
	DPOSConfigTurnBlockCount = 6
	DPOSConfigDelegateCount  = 2
	DPOSConfigDelegateEpoch  = 1
)

type DelegateInfo struct {
	MissedBlocks uint64  `json:"missed_blocks"`
	TotalBlocks  uint64  `json:"total_blocks"`
	Density      float64 `json:"density"`
}

func getSignerAtSlot(delegates []common.Address, slot float64) common.Address {

	if DPOSConfigDelegateCount == 0 || DPOSConfigTurnBlockCount == 0 {
		return common.Address{}
	}

	slot = slot / float64(DPOSConfigTurnBlockCount)
	s := int(slot) % int(DPOSConfigDelegateCount)

	if s < len(delegates) {
		return delegates[s]
	}

	return common.Address{}
}

func getDelegatesStats() (map[string]interface{}, error) {

	dbc := db.GetClient()
	if dbc == nil {
		return nil, errors.New("Failed to open DB")
	}

	// 1. get latest block
	// TODO: get only delegates and timestamp
	latestBlockNumber, err := dbc.GetLatestBlockNumber()
	if err != nil {
		return nil, err
	}

	latestBlock, err := dbc.GetBlockByID(latestBlockNumber)
	if err != nil {
		return nil, err
	}

	totalMissedBlock := 0
	delegatesMap := make(map[common.Address]DelegateInfo, len(latestBlock.Delegates))

	// 2. loop back for `blockDensityLookBackTime` seconds to check for missed blocks by producers
	for i := 0; i < blockDensityLookBackTime; i++ {
		timestamp := float64(latestBlock.TimeStamp) - float64(i)
		slot := timestamp / float64(DPOSConfigPeriod)

		// 3. find the producer who had to produce the block at that time
		origProducer := getSignerAtSlot(latestBlock.Delegates, slot)

		// 4. check if this producer produced the block
		actualProducer, err := dbc.GetBlockProducerAtTimeStamp(hexutil.Uint64(timestamp))
		if err != nil {
			return nil, err
		}

		if _, exists := delegatesMap[origProducer]; !exists {
			delegatesMap[origProducer] = DelegateInfo{
				MissedBlocks: 0,
				TotalBlocks:  0,
				Density:      0,
			}
		}

		delegateInfo := delegatesMap[origProducer]

		delegateInfo.TotalBlocks++

		if *actualProducer != origProducer {
			delegateInfo.MissedBlocks++
			totalMissedBlock++
		}

		delegatesMap[origProducer] = delegateInfo
	}

	// 5. calc density for delegates
	for address, delegateInfo := range delegatesMap {
		delegateInfo.Density = float64(1) - (float64(delegateInfo.MissedBlocks) / float64(delegateInfo.TotalBlocks))

		delegatesMap[address] = delegateInfo
	}

	result := map[string]interface{}{
		// "address": address,
		"total_blocks_examined": blockDensityLookBackTime,
		"total_missed_blocks":   totalMissedBlock,
		"delegates":             latestBlock.Delegates,
		"delegates_info":        delegatesMap,
	}

	return result, nil
}
