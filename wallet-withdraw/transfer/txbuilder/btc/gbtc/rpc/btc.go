package rpc

import (
	"encoding/hex"
	"fmt"
	"math"

	"upex-wallet/wallet-base/jsonrpc"
	"upex-wallet/wallet-withdraw/transfer/txbuilder/btc/gbtc"
)

type BTCRPC struct {
	Client *jsonrpc.Client
}

func NewBTCRPC(url string) gbtc.RPC {
	return newBTCRPC(url)
}

func newBTCRPC(url string) *BTCRPC {
	return &BTCRPC{
		Client: jsonrpc.NewClient(url, jsonrpc.JsonRPCV1),
	}
}

func (r *BTCRPC) GetBestblockhash() (string, error) {
	var (
		hash string
		err  error
	)
	err = r.Client.Call("getbestblockhash", nil, &hash)
	return hash, err
}

// GetBlockByHash returns block information by hash.
func (r *BTCRPC) GetBlockByHash(hash string) ([]byte, error) {
	var (
		blockData []byte
		err       error
	)
	err = r.Client.Call("getblock", jsonrpc.Params{hash}, &blockData)
	return blockData, err
}

// GetBlockByHeight returns block with block height.
func (r *BTCRPC) GetBlockByHeight(height uint64) ([]byte, error) {
	var (
		blockHash string
		err       error
	)

	err = r.Client.Call("getblockhash", jsonrpc.Params{height}, &blockHash)
	if err != nil {
		return nil, err
	}

	return r.GetBlockByHash(blockHash)
}

func (r *BTCRPC) GetRawTransaction(txhash string) (*gbtc.Transaction, error) {
	var (
		txHex string
		err   error
	)

	err = r.Client.Call("getrawtransaction", jsonrpc.Params{txhash}, &txHex)
	if err != nil {
		return nil, err
	}

	raw, err := hex.DecodeString(txHex)
	if err != nil {
		return nil, err
	}

	var tx gbtc.Transaction
	err = tx.SetBytes(raw)
	if err != nil {
		return nil, err
	}

	return &tx, nil
}

// GetTransactionDetail returns raw transaction by transaction hash.
func (r *BTCRPC) GetTransactionDetail(txhash string) ([]byte, error) {
	var (
		tx  []byte
		err error
	)

	err = r.Client.Call("getrawtransaction", jsonrpc.Params{txhash, 1}, &tx)
	return tx, err
}

func (r *BTCRPC) CreateRawTransaction(version uint32, preOuts []*gbtc.OutputPoint, outs []*gbtc.Output) (*gbtc.Transaction, error) {
	var tx gbtc.Transaction
	tx.Version = version
	for _, pre := range preOuts {
		tx.Inputs = append(tx.Inputs, &gbtc.Input{
			PreOutput: pre,
			Sequence:  math.MaxUint32,
		})
	}

	tx.Outputs = outs

	return &tx, nil
}

func (r *BTCRPC) SendRawTransaction(tx *gbtc.Transaction) (string, error) {
	var (
		hexStr = hex.EncodeToString(tx.Bytes())
		hash   string
		err    error
	)

	err = r.Client.Call("sendrawtransaction", jsonrpc.Params{hexStr}, &hash)
	return hash, err
}

func (r *BTCRPC) EstimateSmartFee(confirmNum int) (float64, error) {
	var (
		result struct {
			FeeRate float64  `json:"feerate"`
			Errors  []string `json:"errors"`
			Blocks  int      `json:"blocks"`
		}
		err error
	)

	err = r.Client.Call("estimatesmartfee", jsonrpc.Params{confirmNum}, &result)
	if err != nil {
		return 0, err
	}

	if len(result.Errors) > 0 {
		return 0, fmt.Errorf("%v", result.Errors)
	}

	return result.FeeRate, nil
}
