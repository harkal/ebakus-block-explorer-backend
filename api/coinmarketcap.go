package webapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/urfave/cli"
)

const coinmarketcapEndpoint = "https://pro-api.coinmarketcap.com/v1/tools/price-conversion"
const coinmarketcapEBKCurrencyId = "4778"

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

type PriceConversion struct {
	Data struct {
		Id     uint64
		Symbol string
		Quote  struct {
			USD struct {
				Price float64
			}
		}
	}
}

func (p PriceConversion) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`{"currency":"%s","usd_rate":"%f"}`, p.Data.Symbol, p.Data.Quote.USD.Price)), nil
}

func GetLatestUSDConversionRate() ([]byte, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", coinmarketcapEndpoint, nil)
	if err != nil {
		log.Printf("! Failed to create HTTP request. Error: %s", err.Error())
		return nil, errors.New("Failed to fetch conversion rate")
	}

	q := url.Values{}
	q.Add("id", coinmarketcapEBKCurrencyId)
	q.Add("amount", "1")
	// q.Add("convert", "USD")

	req.Header.Set("Accepts", "application/json")
	req.Header.Add("X-CMC_PRO_API_KEY", coinmarketcapAPIKey)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("! Error sending HTTP request to server: %s", err.Error())
		return nil, errors.New("Failed to fetch conversion rate")
	}

	respBody, _ := ioutil.ReadAll(resp.Body)

	var rate PriceConversion
	json.Unmarshal(respBody, &rate)

	if rate.Data.Symbol != "EBK" {
		log.Printf("! API returned data for wrong coin. (%s)", rate.Data.Symbol)
		return nil, errors.New("Failed to fetch conversion rate")
	}

	return json.Marshal(rate)
}
