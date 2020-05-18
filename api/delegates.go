package webapi

import (
	"errors"
	"sort"

	"github.com/ebakus/ebakus-block-explorer-backend/db"
	"github.com/ebakus/ebakus-block-explorer-backend/ipc"
	"github.com/ebakus/ebakus-block-explorer-backend/models"
	"github.com/ebakus/go-ebakus/common"
	"github.com/ebakus/go-ebakus/common/hexutil"
	"github.com/ebakus/go-ebakus/params"
)

var (
	// intervals to find block density by looking back in seconds from last block
	blockDensityLookBackTimes = []int{
		5 * 60,  // 5 minutes
		60 * 60, // 1 hour
	}
)

var (
	ErrAddressNotFoundInDelegates = errors.New("Address not found in delegates")
)

func getSignerAtSlot(delegates []common.Address, slot float64) common.Address {
	dposConfig := params.MainnetDPOSConfig

	if dposConfig.DelegateCount == 0 || dposConfig.TurnBlockCount == 0 {
		return common.Address{}
	}

	slot = slot / float64(dposConfig.TurnBlockCount)
	s := int(slot) % int(dposConfig.DelegateCount)

	if s < len(delegates) {
		return delegates[s]
	}

	return common.Address{}
}

func getDelegatesStats(address string) (map[string]interface{}, error) {
	dposConfig := params.MainnetDPOSConfig

	isAddressLookup := common.IsHexAddress(address)
	lookupAddress := common.HexToAddress(address)

	// sort blockDensityLookBackTimes for our algorithm
	sort.Ints(blockDensityLookBackTimes)
	longestBlockDensityLookBackTime := blockDensityLookBackTimes[len(blockDensityLookBackTimes)-1]

	dbc := db.GetClient()
	if dbc == nil {
		return nil, errors.New("Failed to open DB")
	}

	ipc := ipc.GetIPC()
	if ipc == nil {
		return nil, errors.New("Failed to find IPC connection")
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

	// if lookupAddress not in delegates then skip
	if isAddressLookup {
		if addressFound, err := ipc.CheckDelegateElected(lookupAddress, -1); !addressFound || err != nil {
			return nil, ErrAddressNotFoundInDelegates
		}
	}

	// get delegate votes from node, only for last block
	delegateVotes, err := ipc.GetDelegates(latestBlockNumber)
	if err != nil {
		return nil, err
	}

	// create a map using `address` as key for our algorithm lookup
	delegateVotesMap := make(map[common.Address]uint64)
	for _, delegateVoteInfo := range delegateVotes {
		delegateVotesMap[delegateVoteInfo.Address] = delegateVoteInfo.Stake
	}

	// 2. get latest blocks from DB during the last `blockDensityLookBackTime` seconds
	timestampOfEarlierBlock := float64(latestBlock.TimeStamp) - float64(longestBlockDensityLookBackTime) - 1 // -1 for reading the delegates of parent block
	latestBlocks, err := dbc.GetBlocksByTimestamp(hexutil.Uint64(timestampOfEarlierBlock), models.TIMESTAMP_GREATER_EQUAL_THAN, "")
	if err != nil {
		return nil, err
	}

	// map to store the end results
	delegatesInfo := make(map[common.Address][]models.DelegateInfo)

	// temp map used during the runtime of our loop
	delegatesRuntime := make(map[common.Address]models.DelegateInfo)

	totalMissedBlocks := 0
	remainingLookBackPeriods := blockDensityLookBackTimes

	parentBlock, latestBlocks := latestBlocks[0], latestBlocks[1:]

	// 3. loop back for `blockDensityLookBackTime` seconds to check for missed blocks by producers
	for idx, block := range latestBlocks {

		if uint64(parentBlock.TimeStamp)-uint64(block.TimeStamp) > 1 {
			totalMissedBlocks += int(uint64(parentBlock.TimeStamp)-uint64(block.TimeStamp)) - 1
		}

		slot := float64(block.TimeStamp) / float64(dposConfig.Period)

		// 4. find the producer who had to produce the block at that time
		origProducer := getSignerAtSlot(parentBlock.Delegates, slot)

		// handle when:
		//   either we search for all delegates
		//   or a specific address and it matches
		if !isAddressLookup || (isAddressLookup && origProducer == lookupAddress) {

			// init DelegateInfo for new delegates
			if _, exists := delegatesRuntime[origProducer]; !exists {
				delegatesRuntime[origProducer] = models.DelegateInfo{
					Address:         origProducer,
					SecondsExamined: 0,
					MissedBlocks:    0,
					TotalBlocks:     0,
					Density:         0,
				}
			}

			delegateInfo := delegatesRuntime[origProducer]
			delegateInfo.TotalBlocks++

			// 5. check if this producer produced the block
			if block.Producer != origProducer {
				delegateInfo.MissedBlocks++
				totalMissedBlocks++
			}

			//  store delegateInfo in the temp runtime map
			delegatesRuntime[origProducer] = delegateInfo
		}

		// check if next lookBack period reached, in order to push delegate info into end result
		if idx == remainingLookBackPeriods[0] {

			// remove existing lookBack period from array and move to next one
			remainingLookBackPeriods = remainingLookBackPeriods[1:]

			// 6. handle data for current period for all delegates
			for curAddress, curDelegateInfo := range delegatesRuntime {
				curDelegateInfo.SecondsExamined = uint64(idx)
				curDelegateInfo.Density = float64(1) - (float64(curDelegateInfo.MissedBlocks) / float64(curDelegateInfo.TotalBlocks))

				// hack: we inject stake into all periods for now, it's not correct
				// as we have stake for latest block only and not for the actual block
				if stake, ok := delegateVotesMap[curAddress]; ok {
					curDelegateInfo.Stake = stake
				}

				// 7. store results for current period for all delegates
				delegatesInfo[curAddress] = append(delegatesInfo[curAddress], curDelegateInfo)
			}
		}

		parentBlock = block
	}

	// parse delegates to array in order to maintain the ordering
	delegatesResponse := make([][]models.DelegateInfo, 0, len(delegatesInfo))
	for _, delegateInfo := range delegatesInfo {
		delegatesResponse = append(delegatesResponse, delegateInfo)
	}

	result := map[string]interface{}{
		"total_seconds_examined": longestBlockDensityLookBackTime,
		"total_missed_blocks":    totalMissedBlocks,
		"delegates":              delegatesResponse,
	}

	if isAddressLookup {
		result["address"] = address
	}

	return result, nil
}
