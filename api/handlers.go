package webapi

import (
	"log"
	"net/http"
	"strconv"

	"bitbucket.org/pantelisss/ebakus_server/models"

	"bitbucket.org/pantelisss/ebakus_server/db"

	"github.com/gorilla/mux"
)

// HandleBlockByID finds and returns block data by id
func HandleBlockByID(w http.ResponseWriter, r *http.Request) {
	dbc := db.GetClient()
	if dbc == nil {
		log.Printf("! Error: DBClient is not initialized!")
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)

	if err != nil {
		log.Printf("! Error: %s", err.Error())
		http.Error(w, "error", http.StatusBadRequest)
		return
	}

	log.Println("Request for:", id)

	if r.Method == "GET" {
		var block *models.Block
		block, err = dbc.GetBlock(id)

		if err != nil {
			log.Printf("! Error: %s", err.Error())
			http.Error(w, "error", http.StatusInternalServerError)
		}

		if block == nil {
			http.Error(w, "error", http.StatusNotFound)
		}

		var res []byte
		res, err = block.MarshalJSON()

		if err != nil {
			log.Printf("! Error: %s", err.Error())
			http.Error(w, "error", http.StatusInternalServerError)
		} else {
			w.Write(res)
		}
	} else {
		http.Error(w, "error", http.StatusBadRequest)
	}
}
