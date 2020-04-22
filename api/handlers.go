package webapi

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strconv"

	"github.com/ebakus/ebakus-block-explorer-backend/db"
	"github.com/ebakus/ebakus-block-explorer-backend/ipc"
	"github.com/ebakus/ebakus-block-explorer-backend/models"
	"github.com/ebakus/ebakus-block-explorer-backend/redis"

	"github.com/ebakus/go-ebakus/common"
	"github.com/ebakus/go-ebakus/params"
	"github.com/gorilla/mux"
)

var (
	valueDecimalPoints = int64(4)
	precisionFactor    = new(big.Int).Exp(big.NewInt(10), big.NewInt(18-valueDecimalPoints), nil)
	ether              = new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
)

// HandleBlockByID finds and returns block data by id
func HandleBlock(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "error", http.StatusBadRequest)
		return
	}

	dbc := db.GetClient()
	if dbc == nil {
		log.Printf("! Error: DBClient is not initialized!")
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)

	var block *models.Block

	if len(vars["param"]) > 2 && vars["param"][1] == 'x' {
		// Case 1: The parameter is Hash
		hash, ok := vars["param"]

		if !ok {
			log.Printf("! Error: %s", errors.New("Parameter is n"))
			http.Error(w, "error", http.StatusBadRequest)
			return
		}

		log.Println("Request Block by Hash:", hash)
		var err error
		block, err = dbc.GetBlockByHash(hash)

		if err != nil {
			log.Printf("! Error: %s", err.Error())
			http.Error(w, "error", http.StatusInternalServerError)
			return
		}

		if block == nil {
			http.Error(w, "error", http.StatusNotFound)
			return
		}
	} else {
		// Case 2: The parameter is ID
		rawId, err := strconv.ParseInt(vars["param"], 10, 64)

		if err != nil {
			log.Printf("! Error: %s", err.Error())
			http.Error(w, "error", http.StatusBadRequest)
			return
		}

		rngParam := r.URL.Query().Get("range")

		if rngParam != "" {
			rng, err := strconv.ParseUint(rngParam, 10, 32)
			if err != nil {
				log.Printf("! Error parsing range: %s", err.Error())
				http.Error(w, "error", http.StatusBadRequest)
				return
			}

			if rawId < 0 && rawId != -1 {
				log.Printf("! Error: Bad negative id")
				http.Error(w, "error", http.StatusBadRequest)
				return
			}

			var id uint32
			if rawId == -1 {
				id = ^uint32(0)
			} else {
				id = uint32(rawId)
			}

			if rng > 100 {
				rng = 100
			}

			blocks, err := dbc.GetBlockRange(id, uint32(rng))

			if err != nil {
				log.Printf("! Error: %s", err.Error())
				http.Error(w, "error", http.StatusInternalServerError)
				return
			}

			if blocks == nil {
				http.Error(w, "error", http.StatusNotFound)
				return
			}

			res, err := json.Marshal(blocks)

			if err != nil {
				log.Printf("! Error: %s", err.Error())
				http.Error(w, "error", http.StatusInternalServerError)
			} else {
				w.Write(res)
			}

			return
		} else {
			id := uint64(rawId)
			log.Println("Request Block by ID:", id)
			block, err = dbc.GetBlockByID(id)

			if err != nil {
				log.Printf("! Error: %s", err.Error())
				http.Error(w, "error", http.StatusInternalServerError)
				return
			}

			if block == nil {
				http.Error(w, "error", http.StatusNotFound)
				return
			}
		}
	}

	res, err := block.MarshalJSON()

	if err != nil {
		log.Printf("! Error: %s", err.Error())
		http.Error(w, "error", http.StatusInternalServerError)
	} else {
		w.Write(res)
	}
}

// HandleTxByHash finds and returns a transaction by hash
func HandleTxByHash(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "error", http.StatusBadRequest)
		return
	}

	dbc := db.GetClient()
	if dbc == nil {
		log.Printf("! Error: DBClient is not initialized!")
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)

	var txf *models.TransactionFull

	hash, ok := vars["hash"]

	if !ok {
		log.Printf("! Error: %s", errors.New("Parameter is n"))
		http.Error(w, "error", http.StatusBadRequest)
		return
	}

	log.Println("Request Transaction by Hash:", hash)
	var err error
	txf, err = dbc.GetTransactionByHash(hash)

	if err != nil {
		log.Printf("! Error: %s", err.Error())
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}

	tx := txf.Tx

	if tx == nil {
		http.Error(w, "error", http.StatusNotFound)
		return
	}

	res, err := txf.MarshalJSON()

	if err != nil {
		log.Printf("! Error: %s", err.Error())
		http.Error(w, "error", http.StatusInternalServerError)
	} else {
		w.Write(res)
	}
}

func HandleAddress(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "error", http.StatusBadRequest)
		return
	}

	dbc := db.GetClient()
	if dbc == nil {
		log.Printf("! Error: DBClient is not initialized!")
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}

	ipc := ipc.GetIPC()
	if ipc == nil {
		log.Printf("! Error: IPCInterface is not initialized!")
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)

	addressHex, ok := vars["address"]
	if !ok || !common.IsHexAddress(addressHex) {
		log.Printf("! Error: %s", errors.New("Parameter is n"))
		http.Error(w, "error", http.StatusBadRequest)
		return
	}

	// correct case sensivity for redis
	address := common.HexToAddress(addressHex)
	addressHex = address.Hex()
	redisKey := "address:" + addressHex

	log.Println("Request Address info for:", addressHex)

	if ok, _ := redis.Exists(redisKey); ok {
		if res, err := redis.Get(redisKey); err == nil {
			w.Write(res)
			return
		}
	}

	blockRewards, txCount, err := dbc.GetAddressTotals(addressHex)
	balance, err := ipc.GetAddressBalance(address)
	stake, err := ipc.GetAddressStaked(address)

	result := models.AddressResult{
		Address:      address,
		Balance:      balance,
		Stake:        stake,
		TxCount:      txCount,
		BlockRewards: blockRewards,
	}

	if addressEns, err := dbc.GetEnsName(addressHex); err == nil {
		result.AddressEns = &addressEns
	}

	res, err := json.Marshal(result)

	if err != nil {
		log.Printf("! Error: %s", err.Error())
		http.Error(w, "error", http.StatusInternalServerError)
	} else {
		redis.Set(redisKey, res)
		redis.Expire(redisKey, 1) // 1 second
		w.Write(res)
	}
}

// HandleTxByAddress finds and returns a transaction by address (from or to)
func HandleTxByAddress(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "error", http.StatusBadRequest)
		return
	}

	dbc := db.GetClient()
	if dbc == nil {
		log.Printf("! Error: DBClient is not initialized!")
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)

	var txs []models.TransactionFull

	address, _ := vars["address"]
	reference, ok := vars["ref"]

	if !ok {
		log.Printf("! Error: %s", errors.New("Parameter is n"))
		http.Error(w, "error", http.StatusBadRequest)
		return
	}

	offsetString := r.URL.Query().Get("offset")
	limitString := r.URL.Query().Get("limit")
	orderString := r.URL.Query().Get("order")

	var offset, limit uint64
	var err error
	if offsetString != "" {
		offset, err = strconv.ParseUint(offsetString, 10, 32)
		if err != nil {
			log.Printf("! Error parsing range: %s", err.Error())
			http.Error(w, "error", http.StatusBadRequest)
			return
		}
	} else {
		offset = 0
	}

	if limitString != "" {
		limit, err = strconv.ParseUint(limitString, 10, 32)
		if err != nil {
			log.Printf("! Error parsing range: %s", err.Error())
			http.Error(w, "error", http.StatusBadRequest)
			return
		}
	} else {
		limit = 20
	}

	if orderString == "" {
		orderString = "asc"
	}

	if orderString != "asc" && orderString != "desc" {
		orderString = "asc"
	}

	log.Println("Request Transaction by Address:", address, "-", reference, offset, limit, orderString)

	switch reference {
	case "from":
		txs, err = dbc.GetTransactionsByAddress(address, models.ADDRESS_FROM, offset, limit, orderString)
	case "to":
		txs, err = dbc.GetTransactionsByAddress(address, models.ADDRESS_TO, offset, limit, orderString)
	case "all":
		txs, err = dbc.GetTransactionsByAddress(address, models.ADDRESS_ALL, offset, limit, orderString)
	case "block":
		txs, err = dbc.GetTransactionsByAddress(address, models.ADDRESS_BLOCKHASH, offset, limit, orderString)
	case "latest":
		txs, err = dbc.GetTransactionsByAddress(address, models.LATEST, offset, limit, orderString)
	default:
		http.Error(w, "error", http.StatusBadRequest)
		return
	}

	if err != nil {
		log.Printf("! Error: %s", err.Error())
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}

	if txs == nil {
		http.Error(w, "error", http.StatusNotFound)
		return
	}

	res, err := json.Marshal(txs)

	if err != nil {
		log.Printf("! Error: %s", err.Error())
		http.Error(w, "error", http.StatusInternalServerError)
	} else {
		w.Write(res)
	}
}

// HandleStats returns stats for producers
func HandleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "error", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	address, ok := vars["address"]
	if ok {
		log.Println("Request Stats for:", address)
	}

	// correct case sensivity for redis
	redisKey := "stats"
	if common.IsHexAddress(address) {
		redisKey += ":" + common.HexToAddress(address).Hex()
	}

	if ok, _ := redis.Exists(redisKey); ok {
		if res, err := redis.Get(redisKey); err == nil {
			w.Write(res)
			return
		}
	}

	result, err := getDelegatesStats(address)
	if err != nil {
		log.Printf("! Error: %s", err.Error())

		if err == ErrAddressNotFoundInDelegates {
			http.Error(w, "error", http.StatusNotFound)
		} else {
			http.Error(w, "error", http.StatusInternalServerError)
		}
		return
	}

	res, err := json.Marshal(result)

	if err != nil {
		log.Printf("! Error: %s", err.Error())
		http.Error(w, "error", http.StatusInternalServerError)
	} else {
		redis.Set(redisKey, res)
		redis.Expire(redisKey, 1)
		w.Write(res)
	}
}

// HandleDelegates returns the delegates for block
func HandleDelegates(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "error", http.StatusBadRequest)
		return
	}

	dbc := db.GetClient()
	if dbc == nil {
		log.Printf("! Error: DBClient is not initialized!")
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}

	ipc := ipc.GetIPC()
	if ipc == nil {
		log.Printf("! Error: IPCInterface is not initialized!")
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)

	redisKey := "delegates"
	redisExpiryTime := uint64(1)

	var blockNumber uint64
	rawId, err := strconv.ParseInt(vars["number"], 10, 64)
	if err == nil {
		// Block number requested
		blockNumber = uint64(rawId)

		log.Println("Request Delegates for block number:", blockNumber)

		if blockNumber < 0 {
			log.Printf("! Error: Bad negative id")
			http.Error(w, "error", http.StatusBadRequest)
			return
		}

		redisKey = fmt.Sprintf("%s:%d", redisKey, blockNumber)
		redisExpiryTime = 60 * 60 * 24 // 1 day
	} else {
		// Latest block requested
		log.Println("Request Delegates for latest block")

		var err error
		blockNumber, err = dbc.GetLatestBlockNumber()
		if err != nil {
			log.Printf("! Error: %s", err.Error())
			http.Error(w, "error", http.StatusInternalServerError)
			return
		}
	}

	if ok, _ := redis.Exists(redisKey); ok {
		if res, err := redis.Get(redisKey); err == nil {
			w.Write(res)
			return
		}
	}

	delegates, err := ipc.GetDelegates(blockNumber)
	if err != nil {
		log.Printf("! Error: %s", err.Error())
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}

	res, err := json.Marshal(delegates)

	if err != nil {
		log.Printf("! Error: %s", err.Error())
		http.Error(w, "error", http.StatusInternalServerError)
	} else {
		redis.Set(redisKey, res)
		redis.Expire(redisKey, redisExpiryTime)
		w.Write(res)
	}
}

// HandleABI returns the ABI for a contract
func HandleABI(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "error", http.StatusBadRequest)
		return
	}

	ipc := ipc.GetIPC()
	if ipc == nil {
		log.Printf("! Error: IPC connection is not initialized!")
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	address, ok := vars["address"]
	if !ok || !common.IsHexAddress(address) {
		log.Println("Request ABI for:", address)
		http.Error(w, "error", http.StatusBadRequest)
		return
	}

	contractAddress := common.HexToAddress(address)
	redisKey := "abi:" + contractAddress.Hex()

	if ok, _ := redis.Exists(redisKey); ok {
		if res, err := redis.Get(redisKey); err == nil {
			w.Write(res)
			return
		}
	}

	abi, err := ipc.GetABIForContract(contractAddress)
	if err != nil {
		log.Printf("! Error: %s", err.Error())
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}

	var res []map[string]interface{}
	if err := json.Unmarshal([]byte(abi), &res); err != nil {
		log.Printf("! Error: %s", err.Error())
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}

	out, err := json.Marshal(res)

	if err != nil {
		log.Printf("! Error: %s", err.Error())
		http.Error(w, "error", http.StatusInternalServerError)
	} else {
		redis.Set(redisKey, out)
		redis.Expire(redisKey, 60*60*24) // 1 day
		w.Write(out)
	}
}

// HandleChainInfo returns useful info for this chain
func HandleChainInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "error", http.StatusBadRequest)
		return
	}

	dbc := db.GetClient()
	if dbc == nil {
		log.Printf("! Error: DBClient is not initialized!")
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	redisKey := "chainInfo"

	if ok, _ := redis.Exists(redisKey); ok {
		if res, err := redis.Get(redisKey); err == nil {
			w.Write(res)
			return
		}
	}

	res := make(map[string]interface{})

	var err error
	latestBlockNumber, err := dbc.GetLatestBlockNumber()
	if err != nil {
		log.Printf("! Error: %s", err.Error())
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}

	latestBlock, err := dbc.GetBlockByID(latestBlockNumber)
	if err != nil {
		log.Printf("! Error: %s", err.Error())
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}

	res["block_number"] = latestBlockNumber
	res["block_timestamp"] = uint64(latestBlock.TimeStamp)
	res["block_hash"] = latestBlock.Hash.Hex()

	dposConfig := params.MainnetDPOSConfig
	blockRewards := 3171 * latestBlockNumber
	totalSupply := dposConfig.InitialDistribution + blockRewards
	totalSupplyWei := new(big.Int).Mul(new(big.Int).SetUint64(totalSupply), ether)

	res["total_supply"] = totalSupplyWei
	res["circulating_supply"] = totalSupplyWei

	out, err := json.Marshal(res)

	if err != nil {
		log.Printf("! Error: %s", err.Error())
		http.Error(w, "error", http.StatusInternalServerError)
	} else {
		redis.Set(redisKey, out)
		redis.Expire(redisKey, 1) // 1 second
		w.Write(out)
	}
}

// HandleRichList returns addresses and its balances
func HandleRichList(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "error", http.StatusBadRequest)
		return
	}

	dbc := db.GetClient()
	if dbc == nil {
		log.Printf("! Error: DBClient is not initialized!")
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	offsetString := r.URL.Query().Get("offset")
	limitString := r.URL.Query().Get("limit")

	var offset, limit uint64
	var err error
	if offsetString != "" {
		offset, err = strconv.ParseUint(offsetString, 10, 32)
		if err != nil {
			log.Printf("! Error parsing range: %s", err.Error())
			http.Error(w, "error", http.StatusBadRequest)
			return
		}
	} else {
		offset = 0
	}

	if limitString != "" {
		limit, err = strconv.ParseUint(limitString, 10, 32)
		if err != nil {
			log.Printf("! Error parsing range: %s", err.Error())
			http.Error(w, "error", http.StatusBadRequest)
			return
		}
	} else {
		limit = 100
	}

	log.Println("Request richlist limit/offset: ", limit, offset)

	richlist, err := dbc.GetTopBalances(limit, offset)
	if err != nil {
		log.Printf("! Error: %s", err.Error())
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}

	res, err := json.Marshal(richlist)

	if err != nil {
		log.Printf("! Error: %s", err.Error())
		http.Error(w, "error", http.StatusInternalServerError)
	} else {
		w.Write(res)
	}
}

// HandleAddReverseRegistrar inserts a new namehash -> name map
func HandleAddReverseRegistrar(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "error", http.StatusBadRequest)
		return
	}

	dbc := db.GetClient()
	if dbc == nil {
		log.Printf("! Error: DBClient is not initialized!")
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var ens models.ENS
	err := decoder.Decode(&ens)
	if err != nil {
		log.Println("Error InsertEns failed", err.Error())
		http.Error(w, "error", http.StatusBadRequest)
		return
	}

	err = dbc.InsertEns(ens)
	if err != nil {
		log.Println("Error InsertEns failed", err.Error())
		http.Error(w, "error", http.StatusBadRequest)
		return
	}

	res, err := json.Marshal(ens)

	if err != nil {
		log.Printf("! Error: %s", err.Error())
		http.Error(w, "error", http.StatusInternalServerError)
	} else {
		w.Write(res)
	}
}

// HandleGetReverseRegistrar returns the address for a namehash
func HandleGetReverseRegistrar(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "error", http.StatusBadRequest)
		return
	}

	dbc := db.GetClient()
	if dbc == nil {
		log.Printf("! Error: DBClient is not initialized!")
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	address, ok := vars["address"]
	if !ok {
		log.Printf("! Error: %s", errors.New("Parameter is address"))
		http.Error(w, "error", http.StatusBadRequest)
		return
	}

	name, err := dbc.GetEnsName(address)
	if err != nil {
		log.Printf("! Error: %s", err.Error())
		if err == sql.ErrNoRows {
			http.Error(w, http.StatusText(404), http.StatusNotFound)
		} else {
			http.Error(w, "error", http.StatusInternalServerError)
		}
		return
	}

	res := make(map[string]interface{})
	res["name"] = name

	out, err := json.Marshal(res)

	if err != nil {
		log.Printf("! Error: %s", err.Error())
		http.Error(w, "error", http.StatusInternalServerError)
	} else {
		w.Write(out)
	}
}
