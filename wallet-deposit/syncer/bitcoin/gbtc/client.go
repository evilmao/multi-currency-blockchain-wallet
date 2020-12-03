package gbtc

import (
	"encoding/hex"

	"upex-wallet/wallet-base/jsonrpc"
)

type Client struct {
	*jsonrpc.Client
}

func NewClient(url string) *Client {
	return &Client{
		Client: jsonrpc.NewClient(url, jsonrpc.JsonRPCV1),
	}
}

// GetBestBlockHash returns the best block hash.
func (rpc *Client) GetBestBlockHash() (string, error) {
	var (
		bestBlockHash string
		err           error
	)

	err = rpc.Call("getbestblockhash", jsonrpc.Params{}, &bestBlockHash)
	return bestBlockHash, err
}

// GetBlockByHash returns block information by hash.
func (rpc *Client) GetBlockByHash(h string) ([]byte, error) {
	var (
		blockData []byte
		err       error
	)
	err = rpc.Call("getblock", jsonrpc.Params{h}, &blockData)
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

	err = rpc.Call("getblockhash", jsonrpc.Params{height}, &blockHash)
	return blockHash, err
}

// GetRawTransaction returns raw transaction by transaction hash.
func (rpc *Client) GetRawTransaction(h string) ([]byte, error) {
	var (
		tx  []byte
		err error
	)

	err = rpc.Call("getrawtransaction", jsonrpc.Params{h, 1}, &tx)
	return tx, err
}

func (rpc *Client) GetRawTransactionData(h string, cfg DeserializeConfig) (*Transaction, error) {
	var (
		txHex string
		err   error
	)

	err = rpc.Call("getrawtransaction", jsonrpc.Params{h}, &txHex)
	if err != nil {
		return nil, err
	}

	raw, err := hex.DecodeString(txHex)
	if err != nil {
		return nil, err
	}

	tx := Transaction{
		DeserConf: cfg,
	}
	err = tx.SetBytes(raw)
	if err != nil {
		return nil, err
	}

	return &tx, nil
}

// SendToAddress sends coin to dest address.
func (rpc *Client) SendToAddress(addr, amount string) (string, error) {
	var (
		txid string
		err  error
	)
	rpc.Call("sendtoaddress", jsonrpc.Params{addr, amount}, &txid)
	return txid, err
}

// OmniListBlockTransactions returns the omnilayer transactions in block.
func (rpc *Client) OmniListBlockTransactions(height int64) ([]byte, error) {
	var (
		blockTxs []byte
		err      error
	)

	err = rpc.Call("omni_listblocktransactions", jsonrpc.Params{height}, &blockTxs)
	return blockTxs, err
}

// OmniGetTransaction returns omnilayer raw transaction.
func (rpc *Client) OmniGetTransaction(h string) ([]byte, error) {
	var (
		omniTx []byte
		err    error
	)
	err = rpc.Call("omni_gettransaction", jsonrpc.Params{h}, &omniTx)
	return omniTx, err
}
