package geos

import (
	"upex-wallet/wallet-base/util"
)

type Client struct {
	url string
}

func NewClient(url string) *Client {
	return &Client{
		url: url,
	}
}

func (c *Client) request(method, path string, data interface{}) ([]byte, error) {
	return util.RestRequest(method, c.url+path, map[string]string{"Accept": "application/json"}, data)
}

// GetInfo returns node info.
func (c *Client) GetInfo() ([]byte, error) {
	return c.request("get", "/v1/chain/get_info", nil)
}

// GetBlock returns block info.
func (c *Client) GetBlock(blockID int64) ([]byte, error) {
	return c.request("post", "/v1/chain/get_block", map[string]int64{
		"block_num_or_id": blockID,
	})
}

// GetActions returns actions of a account.
func (c *Client) GetActions(accountName string, pos int, offset int) ([]byte, error) {
	return c.request("post", "/v1/history/get_actions", map[string]interface{}{
		"account_name": accountName,
		"pos":          pos,
		"offset":       offset,
	})
}

// GetTransaction returns the eos raw transaction.
func (c *Client) GetTransaction(txHash string) ([]byte, error) {
	return c.request("post", "/v1/history/get_transaction", map[string]string{
		"id": txHash,
	})
}
