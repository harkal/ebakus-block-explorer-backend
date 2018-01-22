package main

import (
	_ "ebakus_server/db"
	// "encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter().StrictSlash(true)
	// router.HandleFunc("/blocks", getBlocks).Methods("GET")
	router.HandleFunc("/block/{id}", handleBlock).Methods("GET")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func handleBlock(w http.ResponseWriter, r *http.Request) {
 
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	id := vars["id"]

	log.Println("Request for:", id)

	if r.Method == "GET"{
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