package api

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	// For init resty.
	_ "upex-wallet/wallet-base/util"

	"upex-wallet/wallet-base/newbitx/misclib/log"

	"gopkg.in/resty.v1"
)

type Response interface {
	OK() bool
}

// BaseResponse represents the server response message.
type BaseResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
}

func (r *BaseResponse) OK() bool {
	return r.Status == "ok"
}

// BrokerAPI represents a broker api client.
type BrokerAPI struct {
	url        string
	accessKey  string
	privateKey string
}

// NewBrokerAPI returns a broker api client.
func NewBrokerAPI(url, accessKey, privateKey string) *BrokerAPI {
	return &BrokerAPI{
		url,
		accessKey,
		privateKey,
	}
}

func (api *BrokerAPI) sign(method, url, ts string) string {
	res := md5.Sum([]byte(method + url + ts + api.privateKey))
	return hex.EncodeToString(res[:])
}

func (api *BrokerAPI) request(method, path string, output Response) (err error) {
	if len(api.url) == 0 {
		return fmt.Errorf("broker api url is empty")
	}

	ts := strconv.FormatInt(time.Now().Unix()*1000, 10)
	headers := map[string]string{
		"Accept":              "application/json",
		"FC-ACCESS-TIMESTAMP": ts,
		"FC-ACCESS-SIGNATURE": api.sign(method, path, ts),
		"FC-ACCESS-KEY":       api.accessKey,
	}

	var restResp *resty.Response
	switch method {
	case "GET":
		restResp, err = resty.R().
			SetHeaders(headers).
			Get(api.url + path)
	case "POST":
		restResp, err = resty.R().
			SetHeaders(headers).
			Post(api.url + path)
	default:
		return fmt.Errorf("invalid request method: %s", method)
	}

	if err != nil {
		return fmt.Errorf("request %s failed, %v", path, err)
	}

	if restResp.StatusCode() != http.StatusOK {
		return fmt.Errorf("request %s failed, %s(%d)", path, restResp.Status(), restResp.StatusCode())
	}

	if output == nil {
		return nil
	}

	err = json.Unmarshal(restResp.Body(), output)
	if err != nil {
		log.Errorf("request %s failed, %s", path, string(restResp.Body()))
		return fmt.Errorf("request %s, json unmarshal response failed, %v", path, err)
	}

	if ok := output.OK(); !ok {
		return fmt.Errorf("request %s failed, %s", path, string(restResp.Body()))
	}

	return nil
}

// GetSpecialWithdraws returns special withdraws.
func (api *BrokerAPI) GetSpecialWithdraws() (*BaseResponse, error) {
	var resp BaseResponse
	err := api.request("GET", "/openapi/v1/special_withdraws", &resp)
	return &resp, err
}

// NotifyWithdraw notify withdraws status with txid.
func (api *BrokerAPI) NotifyWithdraw(withdrawID, txid string) (*BaseResponse, error) {
	var resp BaseResponse
	err := api.request("POST",
		fmt.Sprintf("/openapi/v1/special_withdraws/special_withdraw/%s/tx_hash/%s/confirm_callback", withdrawID, txid),
		&resp)
	return &resp, err
}

type CurrencyData struct {
	Code            string `json:"code"`
	Symbol          string `json:"name"`
	DepositEnabled  bool   `json:"deposit_enabled"`
	WithdrawEnabled bool   `json:"withdraw_enabled"`
}

type CurrencyDetail struct {
	BlockchainName    string `json:"block_chain"`
	Symbol            string `json:"currency_name"`
	Confirm           int    `json:"deposit_confirm"`
	MinDepositAmount  string `json:"deposit_min_amount"`
	Address           string `json:"contract_address"`
	Decimal           int    `json:"contract_decimal"`
	WithTag           bool   `json:"label_enabled"`
	Kind              string `json:"kind"`
	MaxWithdrawAmount string `json:"withdraw_single_max_amount"`
}

// AddressKind returns the symbol that c' address is same with.
func (c *CurrencyDetail) AddressKind() string {
	if len(c.Kind) == 0 || c.Kind == "others" {
		return c.Symbol
	}

	return c.Kind
}

// UseAddressOf returns whether c use the same address of the symbol.
func (c *CurrencyDetail) UseAddressOf(symbol string) bool {
	return strings.EqualFold(c.AddressKind(), symbol)
}

// IsToken returns whether c is a token.
func (c *CurrencyDetail) IsToken() bool {
	if len(c.Address) > 0 {
		return true
	}

	if strings.EqualFold(c.Symbol, "usdt") {
		return true
	}

	return false
}

// BelongChainName returns the blockchain name that c belong to.
func (c *CurrencyDetail) BelongChainName() string {
	// Format like: btc_omni, eth_erc20, neo_standard.
	bc := strings.Split(c.BlockchainName, "_")
	if len(bc) >= 2 {
		return bc[0]
	}

	if c.IsToken() {
		return c.AddressKind()
	}

	return c.Symbol
}

// ChainBelongTo returns whether c belong to the blockchain.
func (c *CurrencyDetail) ChainBelongTo(blockchain string) bool {
	return strings.EqualFold(c.BelongChainName(), blockchain)
}

// Valid returns whether data of c is valid.
func (c *CurrencyDetail) Valid() bool {
	if !c.IsToken() {
		return strings.EqualFold(c.Symbol, c.BelongChainName())
	}

	switch {
	case strings.Contains(strings.ToLower(c.BlockchainName), "erc20"):
		return strings.EqualFold(c.AddressKind(), "eth")
	case strings.Contains(strings.ToLower(c.BlockchainName), "omni"):
		return strings.EqualFold(c.AddressKind(), "btc")
	}

	return !strings.EqualFold(c.AddressKind(), c.Symbol)
}

type CurrenciesResponse struct {
	BaseResponse
	Data []*CurrencyData `json:"data"`
}

type CurrencyDetailResponse struct {
	BaseResponse
	Data []*CurrencyDetail `json:"data"`
}

// Currencies gets currencies info.
func (api *BrokerAPI) Currencies() (*CurrenciesResponse, error) {
	var resp CurrenciesResponse
	return &resp, api.request("GET", "/openapi/v2/currencies_with_code", &resp)
}

// CurrencyDetails gets tokens info.
func (api *BrokerAPI) CurrencyDetails() (*CurrencyDetailResponse, error) {
	var resp CurrencyDetailResponse
	return &resp, api.request("GET", "/openapi/v2/currencies/dw", &resp)
}

// ChangeWithdrawStatus changes withdraw status.
// The default value of blockchainName is 'default'.
func (api *BrokerAPI) ChangeWithdrawStatus(currency, blockchainName string, enable bool) (*BaseResponse, error) {
	var resp BaseResponse
	return &resp, api.request("POST", fmt.Sprintf("/openapi/v2/currencies/%s/dw/%s/withdraw/%t", currency, blockchainName, enable), &resp)
}
