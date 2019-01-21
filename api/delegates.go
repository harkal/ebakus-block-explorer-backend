package webapi

import (
	"errors"
	"sort"

	"bitbucket.org/pantelisss/ebakus_server/db"
	"bitbucket.org/pantelisss/ebakus_server/ipc"
	"bitbucket.org/pantelisss/ebakus_server/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

var (
	// intervals to find block density by looking back in seconds from last block
	blockDensityLookBackTimes = []int{
		5 * 60,  // 5 minutes
		60 * 60, // 1 hour
	}
)

// DPOSConfig values on running node
// TODO: get those dynamically
var (
	DPOSConfigPeriod         = 1
	DPOSConfigTurnBlockCount = 6
	DPOSConfigDelegateCount  = 2
	DPOSConfigDelegateEpoch  = 1
)

var (
	ErrAddressNotFoundInDelegates = errors.New("Address not found in delegates")
)

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

func getDelegatesStats(address string) (map[string]interface{}, error) {

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
		addressFound := false
		for _, delegate := range latestBlock.Delegates {
			addressFound = lookupAddress == delegate
			if addressFound {
				break
			}
		}

		if !addressFound {
			return nil, ErrAddressNotFoundInDelegates
		}
	}

	// get delegate votes from node, only for last block
	delegateVotes, err := ipc.GetDelegates(latestBlockNumber)
	if err != nil {
		return nil, err
	}

	// create a map using `address` as key for our algorithm lookup
	delegateVotesMap := make(map[common.Address]uint64, len(delegateVotes))
	for _, delegateVoteInfo := range delegateVotes {
		delegateVotesMap[delegateVoteInfo.Address] = delegateVoteInfo.Stake
	}

	// 2. get latest blocks from DB during the last `blockDensityLookBackTime` seconds
	timestampOfEarlierBlock := float64(latestBlock.TimeStamp) - float64(longestBlockDensityLookBackTime)
	latestBlocks, err := dbc.GetBlocksByTimestamp(hexutil.Uint64(timestampOfEarlierBlock), models.TIMESTAMP_GREATER_EQUAL_THAN, address)
	if err != nil {
		return nil, err
	}

	// create a map using `timestamp` as key for our algorithm lookup
	latestBlocksMap := make(map[uint64]models.Block, len(latestBlocks))
	for _, block := range latestBlocks {
		latestBlocksMap[uint64(block.TimeStamp)] = block
	}

	// map to store the end results
	delegatesInfo := make(map[common.Address][]models.DelegateInfo, len(latestBlock.Delegates))

	// temp map used during the runtime of our loop
	delegatesRuntime := make(map[common.Address]models.DelegateInfo, len(latestBlock.Delegates))

	totalMissedBlocks := 0
	remainingLookBackPeriods := blockDensityLookBackTimes

	// 3. loop back for `blockDensityLookBackTime` seconds to check for missed blocks by producers
	for i := 0; i < longestBlockDensityLookBackTime; i++ {
		timestamp := uint64(latestBlock.TimeStamp) - uint64(i)
		slot := float64(timestamp) / float64(DPOSConfigPeriod)

		// 4. find the producer who had to produce the block at that time
		origProducer := getSignerAtSlot(latestBlock.Delegates, slot)

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
			actualProducer := common.Address{}
			block, blockFound := latestBlocksMap[timestamp]
			if blockFound {
				actualProducer = block.Producer
			}

			if !blockFound || actualProducer != origProducer {
				delegateInfo.MissedBlocks++
				totalMissedBlocks++
			}

			//  store delegateInfo in the temp runtime map
			delegatesRuntime[origProducer] = delegateInfo
		}

		// check if next lookBack period reached, in order to push delegate info into end result
		if i+1 == remainingLookBackPeriods[0] {

			// remove existing lookBack period from array and move to next one
			remainingLookBackPeriods = remainingLookBackPeriods[1:]

			// 6. handle data for current period for all delegates
			for curAddress, curDelegateInfo := range delegatesRuntime {
				curDelegateInfo.SecondsExamined = uint64(i + 1)
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
	}

	// parse delegates to array in order to maintain the ordering
	delegatesResponse := make([][]models.DelegateInfo, 0, len(delegatesInfo))
	for _, delegate := range latestBlock.Delegates {
		if delegateInfo, ok := delegatesInfo[delegate]; ok {
			delegatesResponse = append(delegatesResponse, delegateInfo)
		}
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
