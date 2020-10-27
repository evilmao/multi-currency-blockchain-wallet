package eosio

import (
	"encoding/json"

	"upex-wallet/wallet-base/newbitx/misclib/utils"
)

// API represents EOS api client.
type API struct {
	url string
}

// New returns a new eos api client.
func New(url string) *API {
	return &API{
		url: url,
	}
}

// GetActions returns actions of a account.
func (api *API) GetActions(accountName string, pos int, offset int) ([]byte, error) {
	params := make(map[string]interface{})
	params["account_name"] = accountName
	params["pos"] = pos
	params["offset"] = offset

	req, _ := json.Marshal(params)
	actions, err := utils.RestyPost(string(req), api.url+"/v1/history/get_actions")
	return actions, err
}

// GetInfo returns node info.
func (api *API) GetInfo() ([]byte, error) {
	return utils.RestyGet("", api.url+"/v1/chain/get_info")
}

// GetTransaction returns the eos raw transaction.
func (api *API) GetTransaction(txid string) ([]byte, error) {
	params := make(map[string]interface{})
	params["id"] = txid
	req, _ := json.Marshal(params)
	return utils.RestyPost(string(req), api.url+"/v1/history/get_transaction")
}

func (api *API) Close() {

}
