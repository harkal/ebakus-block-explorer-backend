package web3_dao

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"sync"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"

	"ebakus_server/models"
)

type ResponseBase struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id"`
	Error   *ObjectError    `json:"error,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
}

type ObjectError struct {
	Code    int64       `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func (e *ObjectError) Error() string {
	return e.Message

	// var jsonrpc2ErrorMessages = map[int64]string{
	// 	-32700: "Parse error",
	// 	-32600: "Invalid Request",
	// 	-32601: "Method not found",
	// 	-32602: "Invalid params",
	// 	-32603: "Internal error",
	// 	-32000: "Server error",
	// }
	// fmt.Sprintf("%d (%s) %s\n%v", e.Code, jsonrpc2ErrorMessages[e.Code], e.Message, e.Data)
}

type Client struct {
	url        string
	httpClient *http.Client
	id         int64
	idLock     sync.Mutex
}

type Request struct {
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	ID      int64         `json:"id"`
	Params  []interface{} `json:"params"`
}

var cli *rpc.Client

func init() {
	fmt.Println("init()")

	var err error
	/*** Test with rpc ***/
	cli, err = rpc.Dial("/Users/harkal/ebakus/ebakus.ipc")
	if err != nil {
		log.Fatal("Failed to Dial", err.Error())
	}
}

func newClient(url string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &Client{
		url:        url,
		httpClient: httpClient,
	}
}

//
// Get the top block number
//
func GetBlockNumber() (*big.Int, error) {
	var v hexutil.Big

	err := cli.Call(&v, "eth_blockNumber")
	if err != nil {
		return nil, err
	}

	return v.ToInt(), nil
}

func GetBlock(number *big.Int) (*models.Block, error) {
	var v map[string]*json.RawMessage

	err := cli.Call(&v, "eth_getBlockByNumber", hexutil.EncodeBig(number), true)
	if err != nil {
		return nil, err
	}

	bl := models.NewBlockFromWeb3Map(v)

	return bl, nil
}

// func SyncDatabase() {
// 	blockNumber, err := GetBlockNumber()
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}

// }
