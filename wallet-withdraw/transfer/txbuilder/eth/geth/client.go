package geth

import (
	"fmt"

	"upex-wallet/wallet-base/jsonrpc"

	"github.com/buger/jsonparser"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type Client struct {
	rpcClient *jsonrpc.Client
}

func NewClient(rpcUrl string) *Client {
	rpcClient := jsonrpc.NewClient(rpcUrl, jsonrpc.JsonRPCV2)
	return &Client{rpcClient: rpcClient}
}

func (c *Client) GetBlockByNumber(number uint64) ([]byte, error) {
	var result []byte
	err := c.rpcClient.Call("eth_getBlockByNumber", jsonrpc.Params{toBlockNumArg(number), true}, &result)
	return result, err
}

func (c *Client) GetBlockByHash(hash string) ([]byte, error) {
	var result []byte
	err := c.rpcClient.Call("eth_getBlockByHash", jsonrpc.Params{hash, true}, &result)
	return result, err
}

func (c *Client) GetTransactionByHash(hash string) ([]byte, error) {
	var result []byte
	err := c.rpcClient.Call("eth_getTransactionByHash", jsonrpc.Params{hash}, &result)
	if err != nil {
		return nil, fmt.Errorf("eth_getTransactionByHash failed, %v", err)
	} else if len(result) == 0 {
		err = fmt.Errorf("transaction is nil")
	}
	return result, err
}

func (c *Client) GetLatestBlockNumber() (uint64, error) {
	block, err := c.GetBlockByNumber(0)
	if err != nil {
		return 0, fmt.Errorf("get latest block failed, %v", err)
	}
	if len(block) == 0 {
		return 0, fmt.Errorf("latest block is nil")
	}
	height, err := jsonparser.GetString(block, "number")
	if err != nil {
		return 0, fmt.Errorf("parse number failed, %v", err)
	}
	return hexutil.DecodeUint64(height)
}

func (c *Client) GetTransactionReceipt(hash string) ([]byte, error) {
	var result []byte
	err := c.rpcClient.Call("eth_getTransactionReceipt", jsonrpc.Params{hash}, &result)
	if err != nil {
		return nil, fmt.Errorf("eth_getTransactionReceipt failed, %v", err)
	}
	return result, err
}

func toBlockNumArg(number uint64) string {
	if number == 0 {
		return "latest"
	}
	return hexutil.EncodeUint64(number)
}
