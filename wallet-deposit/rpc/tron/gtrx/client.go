package gtrx

import (
	"fmt"
	"strings"

	"upex-wallet/wallet-base/newbitx/misclib/log"
	"upex-wallet/wallet-base/util"
)

type Client struct {
	BaseURL string
}

// NewClient return trx api client
func NewClient(url string) *Client {
	url = strings.TrimRight(url, "/")
	return &Client{url}
}

func (c *Client) GetAssetIssueListByName(name string) ([]byte, error) {
	param := make(map[string]string)
	param["value"] = name
	url := c.BaseURL + "/wallet/getassetissuelistbyname"
	res, err := util.RestRawPost(url, param)
	if err != nil {
		log.Error("get asset issue by name failed", err.Error())
		return nil, err
	}
	return res, nil
}

func (c *Client) GetSolidityCurrentBlock() ([]byte, error) {
	url := c.BaseURL + "/walletsolidity/getnowblock"
	return util.RestRawPost(url, nil)
}

func (c *Client) GetSolidityBlockByNum(num uint64) ([]byte, error) {
	url := c.BaseURL + "/walletsolidity/getblockbynum"
	params := make(map[string]uint64)
	params["num"] = num
	res, err := util.RestRawPost(url, params)
	if err != nil {
		return nil, fmt.Errorf(" walletsolidity getblockbynum faield,%v", err)
	}
	return res, nil
}

func (c *Client) GetSolidityTransactionById(txid string) ([]byte, error) {
	url := c.BaseURL + "/walletsolidity/gettransactionbyid"
	params := make(map[string]string)
	params["value"] = txid
	res, err := util.RestRawPost(url, params)
	if err != nil {
		return nil, fmt.Errorf(" walletsolidity gettransactionbyid faield,%v", err)
	}
	return res, nil
}

func (c *Client) GetTransactionInfoById(txid string) ([]byte, error) {
	url := c.BaseURL + "/walletsolidity/gettransactioninfobyid"
	params := make(map[string]string)
	params["value"] = txid
	res, err := util.RestRawPost(url, params)
	if err != nil {
		return nil, fmt.Errorf(" walletsolidity gettransactioninfobyid faield,%v", err)
	}
	return res, nil
}
