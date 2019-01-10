package main

import (
	"bytes"
	"html/template"
	"log"
	"os"

	api "bitbucket.org/pantelisss/ebakus_server/api"
	"bitbucket.org/pantelisss/ebakus_server/db"

	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"

	cli "gopkg.in/urfave/cli.v1"
)

type explorerContext struct {
	db     *db.DBClient
	router *mux.Router
}

func (ec explorerContext) initExplorer() cli.BeforeFunc {

	// Part of the init that depends on cmd arguments
	return func(c *cli.Context) error {
		var err error
		err = db.InitFromCli(c)
		if err != nil {
			return err
		}
		ec.db = db.GetClient()

		return nil
	}
}

func (ec explorerContext) startServer() cli.ActionFunc {
	templ, err := template.New("webapi_bindaddr").Parse("{{.Address}}:{{.Port}}")

	if err != nil {
		log.Println(err.Error())
	}

	// Part of the init that depends on cmd arguments
	return func(c *cli.Context) error {
		data := struct {
			Address string
			Port    string
		}{
			c.String("address"),
			c.String("port"),
		}

		buff := new(bytes.Buffer)
		err = templ.Execute(buff, data)

		if err != nil {
			log.Println(err.Error())
		}

		log.Printf("* Ebakus explorer started on http://%s", buff.String())

		ec.router = mux.NewRouter().StrictSlash(true)

		// Setup route handlers...
		ec.router.HandleFunc("/block/{param}", api.HandleBlock).Methods("GET")
		ec.router.HandleFunc("/transaction/{ref:(?:latest)}", api.HandleTxByAddress).Methods("GET")
		ec.router.HandleFunc("/transaction/{hash}", api.HandleTxByHash).Methods("GET")
		ec.router.HandleFunc("/transaction/{ref}/{address}", api.HandleTxByAddress).Methods("GET")

		ec.router.HandleFunc("/address/{address}", api.HandleAddress).Methods("GET")
		ec.router.HandleFunc("/stats", api.HandleStats).Methods("GET")

		handler := cors.Default().Handler(ec.router)
		err = http.ListenAndServe(buff.String(), handler)

		if err != nil {
			log.Println(err.Error())
			return err
		}

		return nil
	}
}

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
			Name:  "address",
			Usage: "Network address to bind",
			Value: "0.0.0.0",
		},
		cli.StringFlag{
			Name:  "port",
			Usage: "Port where the API is served",
			Value: "8080",
		},
		cli.StringFlag{
			Name:  "dbhost",
			Usage: "PostgreSQL database hostname",
			Value: "localhost",
		},
		cli.IntFlag{
			Name:  "dbport",
			Usage: "PostgreSQL database port",
			Value: 5432,
		},
		cli.StringFlag{
			Name:  "dbname",
			Usage: "Database name",
			Value: "ebakus",
		},
		cli.StringFlag{
			Name:  "dbuser",
			Usage: "Database username",
			Value: "ebakus",
		},
		cli.StringFlag{
			Name:  "dbpass",
			Usage: "Database user password",
			Value: "",
		},
	}

	var ctx explorerContext

	app.Before = ctx.initExplorer()
	app.Action = ctx.startServer()

	app.Run(os.Args)
}
