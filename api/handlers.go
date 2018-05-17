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
			rng, err := strconv.ParseUint(rngParam, 10, 64)
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

			id := uint64(rawId)

			if rng > 100 {
				rng = 100
			}

			blocks, err := dbc.GetBlockRange(id, rng)

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

	var tx *models.Transaction

	hash, ok := vars["hash"]

	if !ok {
		log.Printf("! Error: %s", errors.New("Parameter is n"))
		http.Error(w, "error", http.StatusBadRequest)
		return
	}

	log.Println("Request Transaction by Hash:", hash)
	var err error
	tx, err = dbc.GetTransactionByHash(hash)

	if err != nil {
		log.Printf("! Error: %s", err.Error())
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}

	if tx == nil {
		http.Error(w, "error", http.StatusNotFound)
		return
	}

	res, err := tx.MarshalJSON()

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

	var txs []models.Transaction

	address, ok := vars["address"]
	reference, ok := vars["ref"]

	if !ok {
		log.Printf("! Error: %s", errors.New("Parameter is n"))
		http.Error(w, "error", http.StatusBadRequest)
		return
	}

	log.Println("Request Transaction by Address:", address, "-", reference)
	var err error

	switch reference {
	case "from":
		txs, err = dbc.GetTransactionsByAddress(address, models.ADDRESS_FROM)
	case "to":
		txs, err = dbc.GetTransactionsByAddress(address, models.ADDRESS_TO)
	case "block":
		txs, err = dbc.GetTransactionsByAddress(address, models.ADDRESS_BLOCKHASH)
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
