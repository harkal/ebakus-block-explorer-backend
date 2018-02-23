package webapi

import (
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
		}

		if block == nil {
			http.Error(w, "error", http.StatusNotFound)
		}
	} else {
		// Case 2: The parameter is ID
		id, err := strconv.ParseUint(vars["param"], 10, 64)

		if err != nil {
			log.Printf("! Error: %s", err.Error())
			http.Error(w, "error", http.StatusBadRequest)
			return
		}

		log.Println("Request Block by ID:", id)
		block, err = dbc.GetBlockByID(id)

		if err != nil {
			log.Printf("! Error: %s", err.Error())
			http.Error(w, "error", http.StatusInternalServerError)
		}

		if block == nil {
			http.Error(w, "error", http.StatusNotFound)
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
