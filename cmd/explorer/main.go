package main

import (
	"bytes"
	"html/template"
	"log"
	"os"
	"os/user"
	"path/filepath"

	api "github.com/ebakus/ebakus-block-explorer-backend/api"
	"github.com/ebakus/ebakus-block-explorer-backend/db"
	ipcModule "github.com/ebakus/ebakus-block-explorer-backend/ipc"
	"github.com/ebakus/ebakus-block-explorer-backend/redis"

	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"

	"github.com/urfave/cli"
	"github.com/urfave/cli/altsrc"
)

type explorerContext struct {
	db     *db.DBClient
	router *mux.Router
}

func expandHome(path string) string {
	if len(path) >= 2 && path[:2] == "~/" {
		usr, _ := user.Current()
		dir := usr.HomeDir
		path = filepath.Join(dir, path[2:])
	}
	return path
}

func (ec explorerContext) initExplorer() cli.BeforeFunc {
	// Part of the init that depends on cmd arguments
	return func(c *cli.Context) error {
		if err := altsrc.InitInputSourceWithContext(c.App.Flags, altsrc.NewYamlSourceFromFlagFunc("config"))(c); err != nil {
			return err
		}

		var err error
		err = db.InitFromCli(c)
		if err != nil {
			return err
		}
		ec.db = db.GetClient()

		ipcFile := expandHome(c.String("ipc"))
		if _, err := ipcModule.NewIPCInterface(ipcFile); err != nil {
			log.Fatal("Failed to connect to ebakus", err)
		}

		if err := redis.InitFromCli(c); err != nil {
			log.Fatal("Failed to connect to redis", err)
		}
		redis.CleanupHook()

		if err := api.InitCoinmarketcapDefaultsFromCli(c); err != nil {
			log.Println(err)
		}

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
		ec.router.HandleFunc("/stats/{address}", api.HandleStats).Methods("GET")

		ec.router.HandleFunc("/rich-list", api.HandleRichList).Methods("GET")

		ec.router.HandleFunc("/ens", api.HandleAddReverseRegistrar).Methods("POST")
		ec.router.HandleFunc("/ens/{address}", api.HandleGetReverseRegistrar).Methods("GET")

		ec.router.HandleFunc("/delegates", api.HandleDelegates).Methods("GET")
		ec.router.HandleFunc("/delegates/{number}", api.HandleDelegates).Methods("GET")

		ec.router.HandleFunc("/abi/{address}", api.HandleABI).Methods("GET")

		ec.router.HandleFunc("/chain-info", api.HandleChainInfo).Methods("GET")

		ec.router.HandleFunc("/conversion-rate", api.HandleGetConversionRate).Methods("GET")

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
		cli.Author{
			Name:  "Chris Ziogas",
			Email: "ziogaschr@gmail.com",
		},
	}
	app.Copyright = "(c) 2018 Ebakus Team"
	app.Usage = "Web API Server for Ebakus"

	app.Flags = []cli.Flag{
		altsrc.NewStringFlag(cli.StringFlag{
			Name:  "address",
			Usage: "Network address to bind",
			Value: "0.0.0.0",
		}),
		altsrc.NewIntFlag(cli.IntFlag{
			Name:  "port",
			Usage: "Port where the API is served",
			Value: 8080,
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name:  "dbhost",
			Usage: "PostgreSQL database hostname",
			Value: "localhost",
		}),
		altsrc.NewIntFlag(cli.IntFlag{
			Name:  "dbport",
			Usage: "PostgreSQL database port",
			Value: 5432,
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name:  "dbname",
			Usage: "Database name",
			Value: "ebakus",
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name:  "dbuser",
			Usage: "Database username",
			Value: "ebakus",
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name:  "dbpass",
			Usage: "Database user password",
			Value: "",
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name:  "ipc",
			Usage: "The ebakus node to connect to e.g. ./ebakus/ebakus.ipc",
			Value: "~/ebakus/ebakus.ipc",
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name:  "redishost",
			Value: "localhost",
		}),
		altsrc.NewIntFlag(cli.IntFlag{
			Name:  "redisport",
			Value: 6379,
		}),
		altsrc.NewIntFlag(cli.IntFlag{
			Name:  "redispoolsize",
			Value: 10,
		}),
		altsrc.NewIntFlag(cli.IntFlag{
			Name:  "redisdbselect",
			Value: 0,
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name:  "coinmarketcapapikey",
			Value: "",
		}),
		cli.StringFlag{
			Name:  "config",
			Value: "config.yaml",
		},
	}

	var ctx explorerContext

	app.Before = ctx.initExplorer()
	app.Action = ctx.startServer()

	app.Run(os.Args)
}
