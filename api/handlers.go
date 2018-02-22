package webapi

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// HandleBlockByID finds and returns block data by id
func HandleBlockByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	id := vars["id"]

	log.Println("Request for:", id)

	if r.Method == "GET" {
		// if error != nil {
		// 	log.Println(error.Error())
		// 	http.Error(w, error.Error(), http.StatusInternalServerError)
		// 	return
		// }
		fmt.Fprint(w, "success")

	} else {
		http.Error(w, "error", http.StatusBadRequest)
	}
}
