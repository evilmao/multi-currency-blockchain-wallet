package ripple

import (
	"encoding/json"

	"github.com/buger/jsonparser"

	"upex-wallet/wallet-base/newbitx/misclib/rpc"
	"upex-wallet/wallet-base/newbitx/misclib/utils"
)

const (
	// If the specified Amount cannot be sent without spending more than SendMax,
	//reduce the received amount instead of failing outright
	TfPartialPayment = 0x20000
)

type ledgerRequest struct {
	Hash         string `json:"ledger_hash,omitempty"`
	Index        uint64 `json:"ledger_index,omitempty"`
	Accounts     bool   `json:"accounts,omitempty"`
	Full         bool   `json:"full,omitempty"`
	Transactions bool   `json:"transactions,omitempty"`
	Expand       bool   `json:"expand,omitempty"`
	OwnerFunds   bool   `json:"owner_funds,omitempty"`
}

type txRequest struct {
	Hash string `json:"transaction"`
	Hex  bool   `json:"binary"`
}

type accountRequest struct {
	Account  string `json:"account"`
	MinIndex int64  `json:"ledger_index_min,omitempty"`
	MaxIndex int64  `json:"ledger_index_max,omitempty"`
	Binary   bool   `json:"binary,omitempty"`
	Forward  bool   `json:"forward,omitempty"`
	Limit    int    `json:"limit,omitempty"`
}

type RippleRPC struct {
	client *rpc.Client
	url    string
}

func New(rpcClient *rpc.Client) *RippleRPC {
	return &RippleRPC{
		client: rpcClient,
	}
}

// DialHTTP is a wrapper of rpc.DialHTTP.
func DialHTTP(url string) *RippleRPC {
	//rpcClient, err := rpc.DialHTTP(url)
	// if err != nil {
	// 	return nil
	// }
	return &RippleRPC{
		// client: rpcClient,
		url: url,
	}
}

func (rpc RippleRPC) GetCurrentIndex() (uint64, error) {
	var (
		respData []byte
		err      error
	)

	//err = rpc.client.Call("ledger_current", nil, &respData)
	respData, err = utils.RestyPost(`{"method": "ledger_current", "params":[]}`, rpc.url)
	if err != nil {
		return 0, err
	}
	ledgerIndex, _ := jsonparser.GetInt(respData, "result", "ledger_current_index")
	return uint64(ledgerIndex), err
}

func (rpc RippleRPC) GetBlockByIndex(h uint64, isTx bool) ([]byte, error) {
	var (
		blockData []byte
		err       error
	)

	req := ledgerRequest{Index: h}
	if isTx {
		req.Transactions = isTx
	}
	//err = rpc.client.Call("ledger", &req, &blockData)
	reqBytes, _ := json.Marshal(req)

	blockData, err = utils.RestyPost(`{"method": "ledger", "params":[`+string(reqBytes)+`]}`, rpc.url)
	return blockData, err
}

func (rpc RippleRPC) GetBlockByHash(h string, isFull bool) ([]byte, error) {
	var (
		blockData []byte
		err       error
	)

	req := ledgerRequest{Hash: h}
	if isFull {
		req.Full = isFull
	}
	//err = rpc.client.Call("ledger", &req, &blockData)
	reqBytes, _ := json.Marshal(req)
	blockData, err = utils.RestyPost(`{"method": "ledger", "params":[`+string(reqBytes)+`]}`, rpc.url)
	return blockData, err
}

func (rpc RippleRPC) GetRawTransaction(h string) ([]byte, error) {
	var (
		tx  []byte
		err error
	)

	//err = rpc.client.Call("tx", &txRequest{h, false}, &tx)
	reqBytes, _ := json.Marshal(txRequest{h, false})
	tx, err = utils.RestyPost(`{"method": "tx", "params":[`+string(reqBytes)+`]}`, rpc.url)
	return tx, err
}

func (rpc RippleRPC) Close() {
	rpc.client.Close()
}

func (rpc RippleRPC) GetAccountTxs(a string, minIndex uint64) ([]byte, error) {
	reqBytes, _ := json.Marshal(accountRequest{
		Account:  a,
		MinIndex: int64(minIndex),
		MaxIndex: -1,
	})
	return utils.RestyPost(`{"method":"account_tx", "params":[`+string(reqBytes)+`]}`, rpc.url)
}
