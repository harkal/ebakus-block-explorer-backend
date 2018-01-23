package web3_dao

import (
	"log"
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	// "math/big"
	"net/http"
	"sync"

	// "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/kr/pretty"
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

var cli *Client

func init() {
	cli = NewClient("http://localhost:8545", nil)
} 

func NewClient(url string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &Client {
		url:        url,
		httpClient: httpClient,
	}
}

func (c *Client) CallMethod(v interface{}, method string, params ...interface{}) error {
	c.idLock.Lock()

	c.id++

	req := Request{
		JSONRPC: "2.0",
		ID:      c.id,
		Method:  method,
		Params:  params,
	}

	c.idLock.Unlock()

	pretty.Println(req)

	payload, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Post(c.url, "application/json", bytes.NewReader(payload))
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var parsed ResponseBase
	err = json.Unmarshal(body, &parsed)
	if err != nil {
		return err
	}

	if parsed.Error != nil {
		return parsed.Error
	}

	if req.ID != parsed.ID || parsed.JSONRPC != "2.0" {
		return errors.New("Error: JSONRPC 2.0 Specification error")
	}

	pretty.Println(parsed)
	println(string(parsed.Result))

	return json.Unmarshal(parsed.Result, v)
}

func GetBlockNumber() (*hexutil.Big, error) {
	var v hexutil.Big

	err := cli.CallMethod(&v, "eth_blockNumber")
	if err != nil {
		return nil, err
	}

	log.Print("block number is ", v.ToInt())
	return &v, nil
}
