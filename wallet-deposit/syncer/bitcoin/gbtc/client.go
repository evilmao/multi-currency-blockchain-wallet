package gbtc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Client struct {
	url string
}

func NewClient(url string) *Client {
	return &Client{url}
}

func (c *Client) Request(method string, params Params) ([]byte, error) {
	reqData := NewRequest(method, params)
	rawData, err := json.Marshal(reqData)
	if err != nil {
		return nil, fmt.Errorf("json marshal request data failed, %v", err)
	}

	req, err := http.NewRequest("POST", c.url, bytes.NewBuffer(rawData))
	if err != nil {
		return nil, fmt.Errorf("create http request failed, %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send http request failed, %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body failed, %v", err)
	}
	return body, nil
}

func (rpc *Client) Call(method string, params Params, result interface{}) error {
	rawData, err := rpc.Request(method, params)
	if err != nil {
		return err
	}

	var respData Response
	err = json.Unmarshal(rawData, &respData)
	if err != nil {
		return fmt.Errorf("json unmarshal response body failed, %v, %s", err, string(rawData))
	}

	if respData.Error != nil {
		return fmt.Errorf("%v", respData.Error)
	}

	return respData.UnmarshalResult(result)
}

// GetBestBlockHash returns the best block hash.
func (rpc *Client) GetBestBlockHash() (string, error) {
	var (
		bestBlockHash string
		err           error
	)

	err = rpc.Call("getbestblockhash", Params{}, &bestBlockHash)
	return bestBlockHash, err
}

// GetBlockByHash returns block information by hash.
func (rpc *Client) GetBlockByHash(h string) ([]byte, error) {
	var (
		blockData []byte
		err       error
	)
	err = rpc.Call("getblock", Params{h}, &blockData)
	return blockData, err
}

// GetFullBlockByHash returns block full information by hash.
func (rpc *Client) GetFullBlockByHash(h string) ([]byte, error) {
	var (
		blockData []byte
		err       error
	)
	err = rpc.Call("getblock", []interface{}{h, 2}, &blockData)

	return blockData, err
}

// GetBlockByHeight returns block information by height.
func (rpc *Client) GetBlockByHeight(h uint64) ([]byte, error) {
	var (
		blockHash string
		blockData []byte
		err       error
	)
	blockHash, err = rpc.GetBlockHash(h)
	if err != nil {
		return blockData, err
	}
	blockData, err = rpc.GetBlockByHash(blockHash)
	return blockData, err
}

// GetFullBlockByHeight returns block full information by height.
func (rpc *Client) GetFullBlockByHeight(h uint64) ([]byte, error) {
	var (
		blockHash string
		blockData []byte
		err       error
	)
	blockHash, err = rpc.GetBlockHash(h)
	if err != nil {
		return blockData, err
	}
	blockData, err = rpc.GetFullBlockByHash(blockHash)

	return blockData, err
}

// GetBlockHash returns block hash with block height.
func (rpc *Client) GetBlockHash(height uint64) (string, error) {
	var (
		blockHash string
		err       error
	)

	err = rpc.Call("getblockhash", Params{height}, &blockHash)

	return blockHash, err
}

// GetRawTransaction returns raw transaction by transaction hash.
func (rpc *Client) GetRawTransaction(h string) ([]byte, error) {
	var (
		tx  []byte
		err error
	)

	err = rpc.Call("getrawtransaction", Params{h, 1}, &tx)
	return tx, err
}

// SendToAddress sends coin to dest address.
func (rpc *Client) SendToAddress(addr, amount string) (string, error) {
	var (
		txid string
		err  error
	)
	rpc.Call("sendtoaddress", Params{addr, amount}, &txid)
	return txid, err
}

// OmniListBlockTransactions returns the omnilayer transactions in block.
func (rpc *Client) OmniListBlockTransactions(height int64) ([]byte, error) {
	var (
		blockTxs []byte
		err      error
	)

	err = rpc.Call("omni_listblocktransactions", Params{height}, &blockTxs)
	return blockTxs, err
}

// OmniGetTransaction returns omnilayer raw transaction.
func (rpc *Client) OmniGetTransaction(h string) ([]byte, error) {
	var (
		omniTx []byte
		err    error
	)
	err = rpc.Call("omni_gettransaction", Params{h}, &omniTx)
	return omniTx, err
}
