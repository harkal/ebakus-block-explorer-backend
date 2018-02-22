package main

import (
	"os"

	api "bitbucket.org/pantelisss/ebakus_server/api"
	_ "bitbucket.org/pantelisss/ebakus_server/db"
	// "encoding/json"

	"net/http"

	"github.com/gorilla/mux"

	cli "gopkg.in/urfave/cli.v1"
)

func main() {
	app := cli.NewApp()
	app.Name = "Ebakus Blockchain Explorer"
	app.Version = "0.0.1"
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "Harry Kalogirou",
			Email: "harkal@gmail.com",
		},
		cli.Author{
			Name:  "Pantelis Giazitsis",
			Email: "burn665@gmail.com",
		},
		cli.Author{
			Name:  "George Koutsikos",
			Email: "ragecryx@gmail.com",
		},
	}
	app.Copyright = "(c) 2018 Ebakus Team"
	app.Usage = "Web API Server for Ebakus"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "address, a",
			Usage: "Network address to bind",
			Value: "0.0.0.0",
		},
		cli.StringFlag{
			Name:  "port, p",
			Usage: "Port where the API is served",
			Value: "8080",
		},
	}

	app.Action = startServer

	app.Run(os.Args)
}

func startServer(c *cli.Context) error {
	router := mux.NewRouter().StrictSlash(true)
	// router.HandleFunc("/blocks", getBlocks).Methods("GET")
	router.HandleFunc("/block/{id}", api.HandleBlockByID).Methods("GET")
	return http.ListenAndServe(":8080", router)
}
