package rpc

import (
	"encoding/hex"
	"math"
	"strconv"

	"upex-wallet/wallet-base/jsonrpc"
	"upex-wallet/wallet-withdraw/transfer/txbuilder/btc/gbtc"
)

type ETPRPC struct {
	Client *jsonrpc.Client
}

func NewETPRPC(url string) gbtc.RPC {
	return &ETPRPC{
		Client: jsonrpc.NewClient(url, jsonrpc.JsonRPCV2),
	}
}

func (r *ETPRPC) GetBestblockhash() (string, error) {
	var (
		hash string
		err  error
	)
	err = r.Client.Call("getbestblockhash", nil, &hash)
	return hash, err
}

// GetBlockByHash returns block information by hash.
func (r *ETPRPC) GetBlockByHash(hash string) ([]byte, error) {
	var (
		blockData []byte
		err       error
	)
	err = r.Client.Call("getblock", jsonrpc.Params{hash}, &blockData)
	return blockData, err
}

// GetBlockByHeight returns block with block height.
func (r *ETPRPC) GetBlockByHeight(height uint64) ([]byte, error) {
	return r.GetBlockByHash(strconv.Itoa(int(height)))
}

func (r *ETPRPC) GetRawTransaction(txhash string) (*gbtc.Transaction, error) {
	var (
		txHex string
		err   error
	)

	err = r.Client.Call("gettransaction", jsonrpc.Params{txhash, "--json=false"}, &txHex)
	if err != nil {
		return nil, err
	}

	raw, err := hex.DecodeString(txHex)
	if err != nil {
		return nil, err
	}

	tx := gbtc.Transaction{
		DeserConf: gbtc.DeserializeConfig{
			WithAttachment: true,
		},
	}
	err = tx.SetBytes(raw)
	if err != nil {
		return nil, err
	}

	return &tx, nil
}

// GetTransactionDetail returns raw transaction by transaction hash.
func (r *ETPRPC) GetTransactionDetail(txhash string) ([]byte, error) {
	var (
		tx  []byte
		err error
	)

	err = r.Client.Call("gettransaction", jsonrpc.Params{txhash}, &tx)
	return tx, err
}

func (r *ETPRPC) CreateRawTransaction(version uint32, preOuts []*gbtc.OutputPoint, outs []*gbtc.Output) (*gbtc.Transaction, error) {
	tx := gbtc.Transaction{
		DeserConf: gbtc.DeserializeConfig{
			WithAttachment: true,
		},
	}

	tx.Version = version
	for _, pre := range preOuts {
		tx.Inputs = append(tx.Inputs, &gbtc.Input{
			PreOutput: pre,
			Sequence:  math.MaxUint32,
		})
	}

	for _, out := range outs {
		if out.Attachment == nil {
			out.Attachment = &gbtc.Attachment{
				Version: 1,
			}
		}
	}

	tx.Outputs = outs

	return &tx, nil
}

func (r *ETPRPC) SendRawTransaction(tx *gbtc.Transaction) (string, error) {
	var (
		hexStr = hex.EncodeToString(tx.Bytes())
		hash   string
		err    error
	)

	err = r.Client.Call("sendrawtx", jsonrpc.Params{hexStr}, &hash)
	return hash, err
}
