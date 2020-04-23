package webapi

import (
	"errors"

	"github.com/urfave/cli"
)

var (
	coinmarketcapAPIKey string
)

// InitCoinmarketcapDefaultsFromCli is the same as Init but receives it's parameters
// from a Context struct of the cli package (aka from program arguments)
func InitCoinmarketcapDefaultsFromCli(c *cli.Context) error {
	coinmarketcapAPIKey = c.String("coinmarketcapapikey")
	if coinmarketcapAPIKey == "" {
		return errors.New("WARNING. No API key for coinmarketcap provided")
	}
	return nil
}
