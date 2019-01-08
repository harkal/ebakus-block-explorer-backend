package webapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"bitbucket.org/pantelisss/ebakus_server/models"

	"bitbucket.org/pantelisss/ebakus_server/db"

	"github.com/gorilla/mux"
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

	hash, ok := vars["hash"]

	if !ok {
		log.Printf("! Error: %s", errors.New("Parameter is n"))
		http.Error(w, "error", http.StatusBadRequest)
		return
	}

	var res []byte
	var err error

	rngParam := r.URL.Query().Get("range")
	if rngParam != "" {
		rng, err := strconv.ParseUint(rngParam, 10, 32)
		if err != nil {
			log.Printf("! Error parsing range: %s", err.Error())
			http.Error(w, "error", http.StatusBadRequest)
			return
		}

		log.Println("Request Transactions after Hash:", hash)

		if rng > 100 {
			rng = 100
		}

		var txfs []models.TransactionFull
		txfs, err = dbc.GetTransactionRange(hash, uint32(rng))

		if err != nil {
			log.Printf("! Error: %s", err.Error())
			http.Error(w, "error", http.StatusInternalServerError)
			return
		}

		if txfs == nil {
			http.Error(w, "error", http.StatusNotFound)
			return
		}

		res, err = json.Marshal(txfs)
	} else {

		var txf *models.TransactionFull

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

		res, err = txf.MarshalJSON()
	}

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

	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)

	address, ok := vars["address"]

	if !ok {
		log.Printf("! Error: %s", errors.New("Parameter is n"))
		http.Error(w, "error", http.StatusBadRequest)
		return
	}

	log.Println("Request Address info for:", address)

	sumIn, sumOut, miningRewards, countIn, countOut, err := dbc.GetAddressTotals(address)

	result := map[string]interface{}{
		"address":        address,
		"total_in":       sumIn,
		"total_out":      sumOut,
		"count_in":       countIn,
		"count_out":      countOut,
		"mining_rewards": miningRewards,
	}

	res, err := json.Marshal(result)

	if err != nil {
		log.Printf("! Error: %s", err.Error())
		http.Error(w, "error", http.StatusInternalServerError)
	} else {
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
