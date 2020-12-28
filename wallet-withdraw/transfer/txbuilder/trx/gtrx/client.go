package gtrx

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
	"gopkg.in/resty.v1"
)

type NormalArgs map[string]interface{}

type TRC20Req struct {
	ContractAddress  string `json:"contract_address"`
	FunctionSelector string `json:"function_selector"`
	Parameter        string `json:"parameter,omitempty"`
	FeeLimit         int64  `json:"fee_limit"`     // 用户在调用或者创建智能合约时，指定的最高可接受的trx费用消耗，单位SUN（1TRX = 1,000,000SUN）
	CallValue        int    `json:"call_value"`    // 本次调用往合约转账的SUN
	OwnerAddress     string `json:"owner_address"` // 发起deploy contract的账户地址，默认为hexString格式
}

type TransferResult struct {
	Result struct {
		Result bool `json:"result"`
	} `json:"result"`
	ConstantResult []string    `json:"constant_result"`
	Transaction    Transaction `json:"transaction"`
}

type Failed struct {
	Code string `json:"code"`
	Msg  string `json:"message"`
	Err  string `json:"Error"`
}

func (f *Failed) Error() error {
	if len(f.Err) > 0 {
		return fmt.Errorf(f.Err)
	}

	if len(f.Code) == 0 {
		return nil
	}

	m := f.Msg
	if mm, err := hex.DecodeString(f.Msg); err == nil {
		m = string(mm)
	}
	return fmt.Errorf("%s(%s)", f.Code, m)
}

type Client struct {
	BaseURL string
}

// NewClient return lisk api client
func NewClient(url string) *Client {
	url = strings.TrimRight(url, "/")
	return &Client{url}
}

func (c *Client) getResponse(url string, data interface{}) ([]byte, error) {
	url = c.BaseURL + url
	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetBody(data).
		Post(url)
	if err != nil {
		return nil, err
	}

	body := resp.Body()
	var failed Failed
	if err := json.Unmarshal(body, &failed); err == nil {
		if err = failed.Error(); err != nil {
			return nil, err
		}
	}
	return body, nil
}

func (c *Client) CreateTransaction(fromAddr string, toAddr string, amount uint64, assetID int) (*Transaction, error) {
	var err error

	fromAddr, err = AddressToHex(fromAddr)
	if err != nil {
		return nil, fmt.Errorf("invalid from-address format")
	}

	toAddr, err = AddressToHex(toAddr)
	if err != nil {
		return nil, fmt.Errorf("invalid to-address format")
	}

	args := NormalArgs{
		"owner_address": fromAddr,
		"to_address":    toAddr,
		"amount":        amount,
	}

	var rawTx []byte
	if assetID <= 0 {
		rawTx, err = c.getResponse("/wallet/createtransaction", args)
	} else {
		args["asset_name"] = hex.EncodeToString([]byte(strconv.Itoa(assetID)))
		rawTx, err = c.getResponse("/wallet/transferasset", args)
	}

	if err != nil {
		return nil, err
	}

	return JSONUnmarshalTx(rawTx)
}

// GetTransactionSign gets transaction signature. ** dangerous **, only for test.
func (c *Client) GetTransactionSign(tx *Transaction, privateKey []byte) ([]byte, error) {
	return c.getResponse("/wallet/gettransactionsign", NormalArgs{
		"transaction": tx,
		"privateKey":  hex.EncodeToString(privateKey),
	})
}

func (c *Client) BroadcastTransaction(tx *Transaction) ([]byte, error) {
	return c.getResponse("/wallet/broadcasttransaction", tx)
}

// GetTransactionByID gets transaction by id.
func (c *Client) GetTransactionByID(txid string) ([]byte, error) {
	return c.getResponse("/wallet/gettransactionbyid", NormalArgs{
		"value": txid,
	})
}

// GetTransactionInfoByID gets transaction info include fee.
func (c *Client) GetTransactionInfoByID(txid string) ([]byte, error) {
	return c.getResponse("/wallet/gettransactioninfobyid", NormalArgs{
		"value": txid,
	})
}

// GetAccount gets account info.
func (c *Client) GetAccount(address string) ([]byte, error) {
	var err error
	address, err = AddressToHex(address)
	if err != nil {
		return nil, fmt.Errorf("invalid address format")
	}

	return c.getResponse("/wallet/getaccount", NormalArgs{
		"address": address,
	})
}

// GetAccountNet gets account net BandWidth.
func (c *Client) GetAccountNet(address string) ([]byte, error) {
	var err error
	address, err = AddressToHex(address)
	if err != nil {
		return nil, fmt.Errorf("invalid address format")
	}

	return c.getResponse("/wallet/getaccountnet", NormalArgs{
		"address": address,
	})
}

// TRC-20 trigger smart contract
func (c *Client) TriggerSmartContract(rawJson string) ([]byte, error) {
	return c.getResponse("/wallet/triggersmartcontract", rawJson)
}

func CreateTrc20TransferReq(fromAddress, toAddress, contractAddress string, amount decimal.Decimal, precision int32) (*TRC20Req, error) {
	toAddrHex, err := AddressToHex(toAddress)
	if err != nil {
		return nil, err
	}

	fromAddrHex, err := AddressToHex(fromAddress)
	if err != nil {
		return nil, err
	}

	contractAddrHex, err := AddressToHex(contractAddress)
	if err != nil {
		return nil, err
	}

	toAddrHex = strings.Repeat("0", ParamLen-len(toAddrHex)) + toAddrHex
	precisionDec := decimal.New(1, precision)
	// amt := big.NewInt(amount.Mul(precisionDec).IntPart())
	amt := amount.Mul(precisionDec).BigInt()

	amtHex := hex.EncodeToString(amt.Bytes())
	amtHex = strings.Repeat("0", ParamLen-len(amtHex)) + amtHex
	parameter := toAddrHex + amtHex

	// FIX: fee limit is the fixed rate of 1trx = 1,000,000sun, can not use contract precision
	// feeLimit := decimal.New(MaxFeeLimit, 0).Mul(precisionDec)
	feeLimit := decimal.New(MaxFeeLimit, 0).Mul(decimal.New(TRX, 0))

	// Generate a transfer transaction
	return &TRC20Req{
		ContractAddress:  contractAddrHex,
		FunctionSelector: "transfer(address,uint256)",
		Parameter:        parameter,
		FeeLimit:         feeLimit.IntPart(),
		CallValue:        0,
		OwnerAddress:     fromAddrHex,
	}, nil

}
