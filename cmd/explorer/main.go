package main

import (
	"bytes"
	"html/template"
	"log"
	"os"

	api "bitbucket.org/pantelisss/ebakus_server/api"
	"bitbucket.org/pantelisss/ebakus_server/db"
	// "encoding/json"

	"net/http"

	"github.com/gorilla/mux"

	cli "gopkg.in/urfave/cli.v1"
)

type explorerContext struct {
	db     *db.DBClient
	router *mux.Router
}

func createDBClient(c *cli.Context) (*db.DBClient, error) {
	dbname := c.String("dbname")
	dbhost := c.String("dbhost")
	dbport := c.Int("dbport")
	dbuser := c.String("dbuser")
	dbpass := c.String("dbpass")

	return db.NewClient(dbname, dbhost, dbport, dbuser, dbpass)
}

func (ec explorerContext) initExplorer() cli.BeforeFunc {
	ec.router = mux.NewRouter().StrictSlash(true)

	// Setup route handlers...
	ec.router.HandleFunc("/block/{id}", api.HandleBlockByID).Methods("GET")

	// Part of the init that depends on cmd arguments
	return func(c *cli.Context) error {
		var err error
		ec.db, err = db.NewClientByCliArguments(c)
		return err
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

		err = http.ListenAndServe(buff.String(), ec.router)

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
