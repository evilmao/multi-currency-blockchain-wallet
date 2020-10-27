package api

import (
	"encoding/json"
	"errors"
	"fmt"

	"upex-wallet/wallet-base/newbitx/misclib/log"

	"gopkg.in/resty.v1"
)

const (
	// StatusOK represents the api response status.
	StatusOK = iota
)

type CAPI struct {
	ClaimURL string
}

// retry counter
var counter = 0

type Response struct {
	Status int         `json:"status"`
	Msg    string      `json:"msg"`
	Data   interface{} `json:"data"`
}

func NewClaim(cURL string) *CAPI {
	resty.SetLogger(log.GetOutput())
	return &CAPI{
		ClaimURL: cURL,
	}
}

func Post(data interface{}, url string) (int, error) {
	var (
		resp Response
	)
	counter = 0
	_resp, err := resty.R().
		SetHeader("Accept", "application/json").
		SetBody(data).
		Post(url)

	if err != nil || _resp.StatusCode() != 200 {
		return counter, errors.New(_resp.Status())
	}
	json.Unmarshal(_resp.Body(), &resp)
	log.Infof("Rest Post url: %s, resp %v", url, resp)
	if resp.Status == StatusOK {
		return 0, nil
	}
	return counter, fmt.Errorf("api response error %s", _resp)
}

func (api *CAPI) DepositClaimed(data map[string]string) (int, error) {
	return Post(data, api.ClaimURL+"/pushclaim")
}
